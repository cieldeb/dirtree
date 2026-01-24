echo "Formatting go files"
go fmt

echo "Building executable..."
go build

echo "Moving executable to ~/.local/bin"
cp dirtree ~/.local/bin

if [ $? -eq 0 ]; then
  echo "Moved new executable to ~/.local/bin"
  exit 0
else
  echo "Watermarker executable copy failed"
fi
