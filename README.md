# Banner API
This is an API to control display of banners.

# Usage
The core of this api is the banner manager. It displays banners using the `DisplayAppropriateBanner` method. A `GetValidBanners` method is also available if custom display logic is needed. To add new banners into the banner repository, use `AddBanner`.

# Installing dependencies

Requires Go 1.11 or later to run, as it uses go mod for package management.

Run `make install`. Make sure you have Go 1.11 or later installed.

# To Lint
Run `make lint`. You will need to install the tool [golangci-lint](https://github.com/golangci/golangci-lint).

# To Run Tests
The tests can be run using `make test`