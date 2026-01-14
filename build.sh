go build

cp dirtree ~/.local/bin

if [ $? -eq 0 ]; then
  echo "Moved new executable to ~/.local/bin"
  exit 0
else
  echo "Watermarker executable update failed"
fi
