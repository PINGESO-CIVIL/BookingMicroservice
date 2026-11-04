// Package model defines the shared booking domain independently of HTTP and MongoDB.
//
// These types represent domain data, not request bodies or database rows. Services
// must validate them and enforce authorization, policy, and transactional rules.
package model
