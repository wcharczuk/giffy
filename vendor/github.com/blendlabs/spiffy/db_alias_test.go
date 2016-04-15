package spiffy

import (
	"fmt"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestAliases(t *testing.T) {
	oldDefaultAlias := defaultAlias
	defaultAlias = ""
	defer func() {
		defaultAlias = oldDefaultAlias
	}()

	a := assert.New(t)
	config := dbConnectionFromEnvironment()

	CreateDbAlias("test", config)

	gotConn := Alias("test")
	a.Equal(config.Username, gotConn.Username, "Alias(name) should return the correct alias.")

	shouldBeNil := DefaultDb()
	a.Nil(shouldBeNil, "DefaultDb() without an alias should return nil.", fmt.Sprintf("%v", shouldBeNil))

	SetDefaultAlias("test")
	defaultConn := DefaultDb()
	a.NotNil(defaultConn, "DefaultDb() with an alias should return the aliased connection.")
}
