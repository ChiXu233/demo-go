package metaloop

import (
	"github.com/go-resty/resty/v2"
)

var MClient = &MetaLoopClient{}

type MetaLoopClient struct {
	Cli *resty.Client
	Url string
}
