package connlimit

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/mailgun/vulcan/limit"
	"github.com/mailgun/vulcan/limit/connlimit"
	"github.com/mailgun/vulcan/middleware"
	"github.com/mailgun/vulcand/plugin"
)

const Type = "connlimit"

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      Type,
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

// Control simultaneous connections for a location per some variable.
type ConnLimit struct {
	Connections int
	Variable    string // Variable defines how the limiting should be done. e.g. 'client.ip' or 'request.header.X-My-Header'
}

// Returns vulcan library compatible middleware
func (r *ConnLimit) NewMiddleware() (middleware.Middleware, error) {
	mapper, err := limit.VariableToMapper(r.Variable)
	if err != nil {
		return nil, err
	}
	return connlimit.NewConnectionLimiter(mapper, r.Connections)
}

func NewConnLimit(connections int, variable string) (*ConnLimit, error) {
	if _, err := limit.VariableToMapper(variable); err != nil {
		return nil, err
	}
	if connections < 0 {
		return nil, fmt.Errorf("Connections should be > 0, got %d", connections)
	}
	return &ConnLimit{
		Connections: connections,
		Variable:    variable,
	}, nil
}

func (cl *ConnLimit) String() string {
	return fmt.Sprintf("connections=%d, variable=%s", cl.Connections, cl.Variable)
}

func FromOther(c ConnLimit) (plugin.Middleware, error) {
	return NewConnLimit(c.Connections, c.Variable)
}

// Constructs the middleware from the command line
func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return NewConnLimit(c.Int("connections"), c.String("var"))
}

func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{"variable, var", "client.ip", "variable to rate against, e.g. client.ip, request.host or request.header.X-Header"},
		cli.IntFlag{"connections, conns", 1, "amount of simultaneous connections allowed per variable value"},
	}
}
