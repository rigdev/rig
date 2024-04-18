package rbac

const (
	ResourceCapsule        = "capsule"
	ResourceImage          = "image"
	ResourceUser           = "user"
	ResourceGroup          = "group"
	ResourceProject        = "project"
	ResourceRole           = "role"
	ResourceServiceAccount = "serviceaccount"
	ResourceSettings       = "settings"
	ResourceCluster        = "cluster"
	ResourceEnvironment    = "environment"
)

func WithWildcard(resource string) string {
	return resource + "/*"
}

func WithID(resource, id string) string {
	return resource + "/" + id
}

func WithEmpty(resource string) string {
	return resource + "/"
}
