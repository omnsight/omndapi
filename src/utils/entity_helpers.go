package utils

import (
	"slices"

	"github.com/samber/lo"
)

func CheckReadPermission(obj interface{}, userId string, userRoles []string) bool {
	// 1. All read APIs should check if owner or read contains any of user id or user roles.

	type ownerGetter interface {
		GetOwner() string
	}
	type readGetter interface {
		GetRead() []string
	}

	// Check Owner
	if og, ok := obj.(ownerGetter); ok {
		if og.GetOwner() == userId {
			return true
		}
	}

	// Check Read
	if rg, ok := obj.(readGetter); ok {
		readList := rg.GetRead()
		if slices.Contains(readList, userId) {
			return true
		}
		if len(lo.Intersect(readList, userRoles)) > 0 {
			return true
		}
	}

	return false
}

func CheckWritePermission(obj interface{}, userId string, userRoles []string) bool {
	// 3. Update APIs should check if write or owner contains any of user id or user roles

	type ownerGetter interface {
		GetOwner() string
	}
	type writeGetter interface {
		GetWrite() []string
	}

	// Check Owner
	if og, ok := obj.(ownerGetter); ok {
		if og.GetOwner() == userId {
			return true
		}
	}

	// Check Write
	if wg, ok := obj.(writeGetter); ok {
		writeList := wg.GetWrite()
		if slices.Contains(writeList, userId) {
			return true
		}
		if len(lo.Intersect(writeList, userRoles)) > 0 {
			return true
		}
	}

	return false
}

func CheckDeletePermission(obj interface{}, userId string) bool {
	// 4. Delete API is limited to only owner.

	type ownerGetter interface {
		GetOwner() string
	}

	if og, ok := obj.(ownerGetter); ok {
		if og.GetOwner() == userId {
			return true
		}
	}
	return false
}
