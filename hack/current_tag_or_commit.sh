tag=$(git tag --points-at HEAD)
if [ -z "$tag" ]; then
    tag=$(git rev-parse --short HEAD)
else
    # strip "v" prefix form tag
    tag=${tag#v}
fi
 
echo $tag