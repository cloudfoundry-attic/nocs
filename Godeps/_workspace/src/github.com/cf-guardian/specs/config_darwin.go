// +build darwin

package specs

// User specifies specific user and group information for the container's
// main process.
type User struct {
	// UID is the user id.
	UID int32 `json:"uid"`
	// GID is the group id.
	GID int32 `json:"gid"`
	// AdditionalGids are additional group ids set for the container's process.
	AdditionalGids []int32 `json:"additionalGids"`
}
