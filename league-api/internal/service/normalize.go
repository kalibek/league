package service

import "strings"

var kazakhToRu = strings.NewReplacer(
	"Қ", "К", "қ", "к",
	"Ң", "Н", "ң", "н",
	"Ү", "У", "ү", "у",
	"Ө", "О", "ө", "о",
	"Ғ", "Г", "ғ", "г",
	"Ұ", "У", "ұ", "у",
	"Ә", "А", "ә", "а",
	"І", "И", "і", "и",
	"Һ", "Х", "һ", "х",
)

func normalizeName(s string) string { return strings.ToLower(kazakhToRu.Replace(s)) }
