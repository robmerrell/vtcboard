package config

import (
	. "launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type configSuite struct{}

var _ = Suite(&configSuite{})

func (s *configSuite) SetUpSuite(c *C) {
	LoadConfig("testconfig")
}

func (s *configSuite) TestTypedValues(c *C) {
	c.Check(String("testvals.stringval"), Equals, "string")
	c.Check(Int("testvals.intval"), Equals, int64(10))
}
