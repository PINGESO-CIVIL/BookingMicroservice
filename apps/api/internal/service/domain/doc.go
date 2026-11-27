// Package domain defines the shared booking domain independently of HTTP and PostgreSQL.
//
// These types represent domain data, not request bodies or database rows. Services
// must validate them and enforce authorization, policy, and transactional rules.
package domain
