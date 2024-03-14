package option

type OpenWrtOption struct {
	Host     string `required:"true"`
	OS       string `default:"istore"`
	Username string `default:"root"`
	Password string `required:"true"`
}
