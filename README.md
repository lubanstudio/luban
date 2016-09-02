![](public/img/luban-brand.png)

Luban is an on-demand building tasks dispatcher for Go.

## Purpose 

Go already supports cross-compilation but no luck if you uses CGO. This project aims to solve CGO problem by delegating build tasks to any available machines that supports native compilation with given OS, Arch and build tags.

The machine does not have to be owned by you, which means anyone who is interesting on providing free CPU resources can take the build task and contribute to the final artifacts.