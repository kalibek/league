package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"league-api/internal/config"
	"league-api/internal/db"
	"league-api/internal/handler"
	"league-api/internal/middleware"
	"league-api/internal/repository/postgres"
	"league-api/internal/service"
	"league-api/internal/ws"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// Repositories
	userRepo := postgres.NewUserRepo(database)
	oauthRepo := postgres.NewOAuthRepo(database)
	leagueRepo := postgres.NewLeagueRepo(database)
	eventRepo := postgres.NewEventRepo(database)
	groupRepo := postgres.NewGroupRepo(database)
	matchRepo := postgres.NewMatchRepo(database)
	ratingRepo := postgres.NewRatingRepo(database)
	profileRepo := postgres.NewProfileRepo(database)

	// WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Services
	authSvc := service.NewAuthService(database, cfg, userRepo, oauthRepo)
	profileSvc := service.NewProfileService(database, profileRepo, userRepo)
	playerSvc := service.NewPlayerService(userRepo, ratingRepo, eventRepo, groupRepo, matchRepo, profileSvc)
	leagueSvc := service.NewLeagueService(database, leagueRepo, userRepo)
	eventSvc := service.NewEventService(database, eventRepo, groupRepo, matchRepo, userRepo)
	ratingSvc := service.NewRatingService(database, userRepo, groupRepo, matchRepo, ratingRepo, eventRepo)
	matchSvc := service.NewMatchService(database, matchRepo, groupRepo, hub)
	groupSvc := service.NewGroupService(database, groupRepo, matchRepo, eventRepo, hub)
	draftSvc := service.NewDraftService(database, leagueRepo, eventRepo, groupRepo, matchRepo, matchSvc, ratingSvc, groupSvc, hub)

	// Handlers
	adminH := handler.NewAdminHandler(ratingSvc)
	authH := handler.NewAuthHandler(authSvc, cfg.FrontendURL)
	playersH := handler.NewPlayersHandler(playerSvc)
	profileH := handler.NewProfileHandler(profileSvc)
	leaguesH := handler.NewLeaguesHandler(leagueSvc)
	eventsH := handler.NewEventsHandler(eventSvc, draftSvc, leagueSvc)
	groupsH := handler.NewGroupsHandler(groupSvc, draftSvc, matchSvc, leagueRepo, eventRepo)
	matchesH := handler.NewMatchesHandler(matchSvc, groupSvc, leagueRepo, eventRepo)
	wsH := handler.NewWebSocketHandler(hub, eventRepo)

	// Router
	r := gin.Default()
	r.Use(middleware.CORS(cfg.FrontendURL))

	authMiddleware := middleware.Auth(authSvc, leagueRepo)

	// Auth routes
	auth := r.Group("/api/v1/auth")
	{
		auth.GET("/login", authH.Login)
		auth.GET("/callback", authH.Callback)
		auth.POST("/logout", authH.Logout)
		auth.GET("/me", authMiddleware, authH.Me)
		auth.POST("/register", authH.Register)
		auth.POST("/login/email", authH.EmailLogin)
	}

	// Public API — no authentication required
	public := r.Group("/api/v1/public")
	{
		public.GET("/players", playersH.List)
		public.GET("/players/:id", playersH.Get)
		public.GET("/players/:id/events", playersH.ListEvents)

		public.GET("/countries", profileH.ListCountries)
		public.GET("/countries/:cid/cities", profileH.ListCities)
		public.GET("/blades", profileH.ListBlades)
		public.GET("/rubbers", profileH.ListRubbers)

		public.GET("/leagues", leaguesH.List)
		public.GET("/leagues/:id", leaguesH.Get)
		public.GET("/leagues/:id/events", eventsH.List)
		public.GET("/leagues/:id/events/:eid", eventsH.Get)

		public.GET("/events/:eid/groups", groupsH.List)
		public.GET("/events/:eid/groups/:gid", groupsH.Get)
		public.GET("/events/:eid/tables-in-use", matchesH.GetTablesInUse)
	}

	// Secured API — authentication required
	secured := r.Group("/api/v1/secured", authMiddleware)
	{
		secured.GET("/profile", profileH.GetMyProfile)
		secured.PUT("/profile", profileH.UpsertMyProfile)
		secured.POST("/countries/:cid/cities", profileH.AddCity)
		secured.POST("/blades", profileH.AddBlade)
		secured.POST("/rubbers", profileH.AddRubber)

		secured.POST("/players", playersH.Create)
		secured.POST("/players/import", playersH.Import)

		secured.POST("/leagues", leaguesH.Create)
		secured.PUT("/leagues/:id/config", leaguesH.UpdateConfig)
		secured.GET("/leagues/:id/roles", leaguesH.GetRoles)
		secured.POST("/leagues/:id/roles", leaguesH.AssignRole)
		secured.DELETE("/leagues/:id/roles", leaguesH.RemoveRole)
		secured.POST("/leagues/:id/events", eventsH.Create)
		secured.POST("/leagues/:id/events/:eid/start", eventsH.Start)
		secured.POST("/leagues/:id/events/:eid/finish", eventsH.Finish)
		secured.POST("/leagues/:id/events/:eid/next", eventsH.CreateNext)
		secured.PUT("/leagues/:id/events/:eid/config", eventsH.UpdateConfig)

		secured.POST("/events/:eid/groups", groupsH.Create)
		secured.POST("/events/:eid/groups/:gid/finish", groupsH.Finish)
		secured.POST("/events/:eid/groups/:gid/reopen", groupsH.Reopen)
		secured.POST("/events/:eid/groups/:gid/players", groupsH.AddPlayer)
		secured.POST("/events/:eid/groups/:gid/add-player", groupsH.AddPlayerToActiveGroup)
		secured.POST("/events/:eid/groups/:gid/seed", groupsH.SeedPlayer)
		secured.DELETE("/events/:eid/groups/:gid/players/:gpid", groupsH.RemovePlayer)
		secured.PUT("/events/:eid/groups/:gid/players/:pid/place", groupsH.SetManualPlace)
		secured.PUT("/events/:eid/groups/:gid/players/:pid/status", groupsH.SetPlayerStatus)
		secured.PUT("/events/:eid/groups/:gid/placement", groupsH.SetManualPlacements)

		secured.PUT("/groups/:gid/matches/:mid", matchesH.UpdateScore)
		secured.DELETE("/groups/:gid/matches/:mid/score", matchesH.ResetScore)
		secured.PUT("/groups/:gid/matches/:mid/table", matchesH.SetTableNumber)
	}

	// Admin API — admin flag required
	admin := r.Group("/api/v1/admin", authMiddleware, middleware.RequireAdmin())
	{
		admin.POST("/ratings/recalculate", adminH.RecalculateRatings)
	}

	// WebSocket
	r.GET("/ws/events/:eid", wsH.Handle)

	log.Printf("server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
