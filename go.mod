module github.com/macroblock/imed

go 1.15

//replace github.com/macroblock/rtimg => ../../macroblock/rtimg

//replace golang.com/x/ => ../../golang.com/x/

require (
	github.com/atotto/clipboard v0.1.4
	github.com/d5/tengo/v2 v2.8.0
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213
	github.com/macroblock/rtimg v0.0.0-20210707074111-12be9d0e886a
	github.com/malashin/ffinfo v0.0.0-20210606231020-f15065768ba1
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/text v0.3.7-0.20210524175448-3115f89c4b99
)
