tag=$(git describe --tags --abbrev=0)
tag=${tag#v}

echo $tag