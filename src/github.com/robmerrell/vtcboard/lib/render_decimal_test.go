package lib

import (
	. "launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type renderDecimalSuite struct{}

var _ = Suite(&renderDecimalSuite{})

func (s *renderDecimalSuite) TestTrailingZeroes(c *C) {
	rendered := RenderIntegerFromString("", "1500500")
	c.Check(rendered, Equals, "1,500,500")
}
