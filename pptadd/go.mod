module github.com/ktye/pptx/pptadd

go 1.13

require (
	github.com/ktye/plot v0.0.0
	github.com/ktye/pptx v0.0.0
	github.com/ktye/pptx/pptxt v0.0.0
)

replace (
	github.com/ktye/plot => ../../plot
	github.com/ktye/pptx => ../
	github.com/ktye/pptx/pptxt => ../pptxt
)
