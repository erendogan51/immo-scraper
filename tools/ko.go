package tools

//go:generate LAST_TAG=$(./hack/most_recent_tag.sh) BUILD_TIME='$(date)' GIT_COMMIT=$(./hack/current_tag_or_commit.sh) go run github.com/google/ko@latest build --local ../cmd/receiver
