# HCL Funcs

This library is a facade that combines all [built-in functions](https://developer.hashicorp.com/packer/docs/templates/hcl_templates/functions) provided by [HashiCorp Packer](https://packer.io/) and some functions from [HashiCorp Terraform](https://www.terraform.io/).

All copied functions are copied from Terraform and Packer's latest MPL 2.0 license version, all referenced functions are based on MPL 2.0 or MIT license.

## Goroutine-local `env` function

`env` function is different than the Packer version, we provided a goroutine-local cache so the caller can set different environment variables for different goroutines, this is very handy when you allow users to set different environment variables for a specified HCL block, like [this example](https://github.com/Azure/grept/blob/main/doc/f/local_shell.md#example). Please check out [this unit test](https://github.com/lonegunmanb/hclfuncs/blob/main/functions_test.go#L27-L61) for details.