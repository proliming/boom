package main

import "errors"

var (
    errNilBoomOpts = errors.New("nil BoomOptions, must specified -u")
    errBoomOpts = errors.New("Invalid BoomOptions.")
    errInvalidHttpMethod = errors.New("Invalid http method.")
    errZeroRate = errors.New("rate must be bigger than zero")
    errBadCert  = errors.New("bad certificate")
)

