package service

// This file adds stub implementations for newly added repository interface methods
// to all existing mock structs across the service test suite.

import (
	"context"
	"strings"

	"league-api/internal/model"
)

// ---- UserRepository new methods ----

func (m *authMockUserRepo) FindAllActive(_ context.Context) ([]model.User, error) {
	return nil, nil
}
func (m *authMockUserRepo) SoftDeleteMerged(_ context.Context, _, _ int64) error { return nil }
func (m *authMockUserRepo) UpdateRatingHistory(_ context.Context, _, _ int64) error { return nil }
func (m *authMockUserRepo) CountPlayers(_ context.Context, _ string) (int, error) { return 0, nil }

func (m evtMockUserRepo) FindAllActive(_ context.Context) ([]model.User, error) { return nil, nil }
func (m evtMockUserRepo) SoftDeleteMerged(_ context.Context, _, _ int64) error  { return nil }
func (m evtMockUserRepo) UpdateRatingHistory(_ context.Context, _, _ int64) error { return nil }
func (m evtMockUserRepo) CountPlayers(_ context.Context, _ string) (int, error)  { return 0, nil }

func (m *mockUserRepoForLeague) FindAllActive(_ context.Context) ([]model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForLeague) SoftDeleteMerged(_ context.Context, _, _ int64) error { return nil }
func (m *mockUserRepoForLeague) UpdateRatingHistory(_ context.Context, _, _ int64) error {
	return nil
}
func (m *mockUserRepoForLeague) CountPlayers(_ context.Context, _ string) (int, error) { return 0, nil }

func (m *mockUserRepoForProfile) FindAllActive(_ context.Context) ([]model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForProfile) SoftDeleteMerged(_ context.Context, _, _ int64) error { return nil }
func (m *mockUserRepoForProfile) UpdateRatingHistory(_ context.Context, _, _ int64) error {
	return nil
}
func (m *mockUserRepoForProfile) CountPlayers(_ context.Context, _ string) (int, error) { return 0, nil }

func (m *ratingMockUserRepo) FindAllActive(_ context.Context) ([]model.User, error) {
	return nil, nil
}
func (m *ratingMockUserRepo) SoftDeleteMerged(_ context.Context, _, _ int64) error { return nil }
func (m *ratingMockUserRepo) UpdateRatingHistory(_ context.Context, _, _ int64) error { return nil }
func (m *ratingMockUserRepo) CountPlayers(_ context.Context, _ string) (int, error) { return 0, nil }

// ---- GroupRepository new methods ----

func groupConflictStub(_ context.Context, _, _ int64) ([]int64, error) { return nil, nil }
func groupSetStatusByUserStub(_ context.Context, _, _ int64, _ model.PlayerStatus) error {
	return nil
}
func groupUpdateGPUserIDStub(_ context.Context, _, _ int64, _ []int64) error { return nil }
func groupDeleteProfileStub(_ context.Context, _ int64) error                { return nil }

func (m *matchSvcMockGroupRepo) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *matchSvcMockGroupRepo) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *matchSvcMockGroupRepo) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *matchSvcMockGroupRepo) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *evtMockGroupRepo) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *evtMockGroupRepo) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *evtMockGroupRepo) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *evtMockGroupRepo) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *draftMockGroupRepo) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *draftMockGroupRepo) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *draftMockGroupRepo) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *draftMockGroupRepo) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *mockGroupRepo) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *mockGroupRepo) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *mockGroupRepo) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *mockGroupRepo) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *mockGroupRepoWithAddCapture) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *mockGroupRepoWithAddCapture) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *mockGroupRepoWithAddCapture) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *mockGroupRepoWithAddCapture) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *mockGroupRepoWithSetStatus) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *mockGroupRepoWithSetStatus) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *mockGroupRepoWithSetStatus) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *mockGroupRepoWithSetStatus) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *mockGroupRepoWithDuplicate) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *mockGroupRepoWithDuplicate) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *mockGroupRepoWithDuplicate) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *mockGroupRepoWithDuplicate) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *mockGroupRepoForActive) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *mockGroupRepoForActive) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *mockGroupRepoForActive) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *mockGroupRepoForActive) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

func (m *ratingMockGroupRepo) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *ratingMockGroupRepo) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *ratingMockGroupRepo) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *ratingMockGroupRepo) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

// ---- RatingRepository new methods ----

func (m *ratingMockRatingRepo) DeleteFromEvent(_ context.Context, _ int64) error { return nil }
func (m *ratingMockRatingRepo) GetLastRatingsBeforeEvent(_ context.Context, _ int64) ([]model.UserRatingSnapshot, error) {
	return nil, nil
}
func (m *ratingMockRatingRepo) GetEarliestEventIDForUser(_ context.Context, _ int64) (int64, bool, error) {
	return 0, false, nil
}

// ---- EventRepository new methods ----

func (m *draftMockEventRepo) ListDoneFromEvent(_ context.Context, _ int64) ([]model.LeagueEvent, error) {
	return nil, nil
}
func (m *evtMockEventRepo) ListDoneFromEvent(_ context.Context, _ int64) ([]model.LeagueEvent, error) {
	return nil, nil
}
func (m *mockEventRepo) ListDoneFromEvent(_ context.Context, _ int64) ([]model.LeagueEvent, error) {
	return nil, nil
}
func (m *mockDraftEventRepo) ListDoneFromEvent(_ context.Context, _ int64) ([]model.LeagueEvent, error) {
	return nil, nil
}
func (m *ratingMockEventRepo) ListDoneFromEvent(_ context.Context, _ int64) ([]model.LeagueEvent, error) {
	return nil, nil
}
func (m *fullEventRepo) ListDoneFromEvent(_ context.Context, _ int64) ([]model.LeagueEvent, error) {
	return nil, nil
}

// ---- RatingService new methods ----

func (m *nopRatingService) RecalculateFromEvent(_ context.Context, _ int64) (RecalcResult, error) {
	return RecalcResult{}, nil
}
func (m *draftMockRatingSvc) RecalculateFromEvent(_ context.Context, _ int64) (RecalcResult, error) {
	return RecalcResult{}, nil
}

// ---- player_service_test.go mocks ----

func (m *psUserRepo) FindAllActive(_ context.Context) ([]model.User, error)          { return nil, nil }
func (m *psUserRepo) SoftDeleteMerged(_ context.Context, _, _ int64) error           { return nil }
func (m *psUserRepo) UpdateRatingHistory(_ context.Context, _, _ int64) error        { return nil }
func (m *psUserRepo) CountPlayers(_ context.Context, q string) (int, error) {
	if q == "" {
		return len(m.users), nil
	}
	var count int
	for _, u := range m.users {
		if strings.Contains(u.FirstName, q) || strings.Contains(u.LastName, q) {
			count++
		}
	}
	return count, nil
}

// ---- player_merge_test.go mocks ----

func (m *mergeUserRepo) CountPlayers(_ context.Context, q string) (int, error) {
	if q == "" {
		return len(m.users), nil
	}
	var count int
	for _, u := range m.users {
		if strings.Contains(u.FirstName, q) || strings.Contains(u.LastName, q) {
			count++
		}
	}
	return count, nil
}

func (m *psRatingRepo) DeleteFromEvent(_ context.Context, _ int64) error { return nil }
func (m *psRatingRepo) GetLastRatingsBeforeEvent(_ context.Context, _ int64) ([]model.UserRatingSnapshot, error) {
	return nil, nil
}
func (m *psRatingRepo) GetEarliestEventIDForUser(_ context.Context, _ int64) (int64, bool, error) {
	return 0, false, nil
}

func (m *psEventRepo) ListDoneFromEvent(_ context.Context, _ int64) ([]model.LeagueEvent, error) {
	return nil, nil
}

func (m *playerTestGroupRepo) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *playerTestGroupRepo) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *playerTestGroupRepo) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *playerTestGroupRepo) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}

// ---- player_service_extra_test.go mocks ----

func (m *playerEventRatingRepo) DeleteFromEvent(_ context.Context, _ int64) error { return nil }
func (m *playerEventRatingRepo) GetLastRatingsBeforeEvent(_ context.Context, _ int64) ([]model.UserRatingSnapshot, error) {
	return nil, nil
}
func (m *playerEventRatingRepo) GetEarliestEventIDForUser(_ context.Context, _ int64) (int64, bool, error) {
	return 0, false, nil
}

// ---- player_service_test.go psGroupRepo stubs ----

func (m *psGroupRepo) FindConflictingGroupIDs(ctx context.Context, s, t int64) ([]int64, error) {
	return groupConflictStub(ctx, s, t)
}
func (m *psGroupRepo) SetPlayerStatusByUser(ctx context.Context, g, u int64, st model.PlayerStatus) error {
	return groupSetStatusByUserStub(ctx, g, u, st)
}
func (m *psGroupRepo) UpdateGroupPlayerUserID(ctx context.Context, s, t int64, ex []int64) error {
	return groupUpdateGPUserIDStub(ctx, s, t, ex)
}
func (m *psGroupRepo) DeletePlayerProfileByUser(ctx context.Context, u int64) error {
	return groupDeleteProfileStub(ctx, u)
}
