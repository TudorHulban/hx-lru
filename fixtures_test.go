package hxlru

import "github.com/TudorHulban/epochid"

type paramsTestLRU struct {
	ProjectID      epochid.EpochID
	ProjectMembers uint8
}

type member struct {
	Name   string
	Skills []string
}
