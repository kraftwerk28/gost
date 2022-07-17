command! -nargs=0 GenDoc
       \ !go run gen_doc.go -t template.md -o readme-test.md blocks/*.go
