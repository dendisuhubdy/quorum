package utils

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/node"
	"github.com/stretchr/testify/assert"
	"gopkg.in/urfave/cli.v1"
)

func TestSetPlugins_whenPluginsNotEnabled(t *testing.T) {
	arbitraryNodeConfig := &node.Config{}
	arbitraryCLIContext := cli.NewContext(nil, &flag.FlagSet{}, nil)

	assert.NoError(t, setPlugins(arbitraryCLIContext, arbitraryNodeConfig))

	assert.Nil(t, arbitraryNodeConfig.Plugins)
}

func TestSetPlugins_whenInvalidFlagsCombination(t *testing.T) {
	arbitraryNodeConfig := &node.Config{}
	fs := &flag.FlagSet{}
	fs.String(PluginSettingsFlag.Name, "", "")
	fs.Bool(PluginSkipVerifyFlag.Name, true, "")
	fs.Bool(PluginLocalVerifyFlag.Name, true, "")
	fs.String(PluginPublicKeyFlag.Name, "", "")
	arbitraryCLIContext := cli.NewContext(nil, fs, nil)
	assert.NoError(t, arbitraryCLIContext.GlobalSet(PluginSettingsFlag.Name, "arbitrary value"))

	verifyErrorMessage(t, arbitraryCLIContext, arbitraryNodeConfig, "only --plugins.skipverify or --plugins.localverify must be set")

	assert.NoError(t, arbitraryCLIContext.GlobalSet(PluginSkipVerifyFlag.Name, "false"))
	assert.NoError(t, arbitraryCLIContext.GlobalSet(PluginLocalVerifyFlag.Name, "false"))
	assert.NoError(t, arbitraryCLIContext.GlobalSet(PluginPublicKeyFlag.Name, "arbitry value"))

	verifyErrorMessage(t, arbitraryCLIContext, arbitraryNodeConfig, "--plugins.localverify is required for setting --plugins.publickey")
}

func TestSetPlugins_whenInvalidPluginSettingsURL(t *testing.T) {
	arbitraryNodeConfig := &node.Config{}
	fs := &flag.FlagSet{}
	fs.String(PluginSettingsFlag.Name, "", "")
	arbitraryCLIContext := cli.NewContext(nil, fs, nil)
	assert.NoError(t, arbitraryCLIContext.GlobalSet(PluginSettingsFlag.Name, "arbitrary value"))

	verifyErrorMessage(t, arbitraryCLIContext, arbitraryNodeConfig, "plugins: unable to create reader due to unsupported scheme ")
}

func TestSetImmutabilityThreshold (t *testing.T) {
	fs := &flag.FlagSet{}
	fs.Int(QuorumImmutabilityThreshold.Name, 0, "")
	arbitraryCLIContext := cli.NewContext(nil, fs, nil)
	assert.NoError(t, arbitraryCLIContext.GlobalSet(QuorumImmutabilityThreshold.Name, strconv.Itoa(100000)))
	assert.True(t, arbitraryCLIContext.GlobalIsSet(QuorumImmutabilityThreshold.Name) == true, "immutability threshold flag not set" )
	assert.True(t,arbitraryCLIContext.GlobalInt(QuorumImmutabilityThreshold.Name) == 100000, "immutability threshold value not set")
}

func TestSetPlugins_whenTypical(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "q-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	arbitraryJSONFile := path.Join(tmpDir, "arbitary.json")
	if err := ioutil.WriteFile(arbitraryJSONFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	arbitraryNodeConfig := &node.Config{}
	fs := &flag.FlagSet{}
	fs.String(PluginSettingsFlag.Name, "", "")
	arbitraryCLIContext := cli.NewContext(nil, fs, nil)
	assert.NoError(t, arbitraryCLIContext.GlobalSet(PluginSettingsFlag.Name, "file://"+arbitraryJSONFile))

	assert.NoError(t, setPlugins(arbitraryCLIContext, arbitraryNodeConfig))

	assert.NotNil(t, arbitraryNodeConfig.Plugins)
}

func verifyErrorMessage(t *testing.T, ctx *cli.Context, cfg *node.Config, expectedMsg string) {
	err := setPlugins(ctx, cfg)
	assert.EqualError(t, err, expectedMsg)
}

func Test_SplitTagsFlag(t *testing.T) {
	tests := []struct {
		name string
		args string
		want map[string]string
	}{
		{
			"2 tags case",
			"host=localhost,bzzkey=123",
			map[string]string{
				"host":   "localhost",
				"bzzkey": "123",
			},
		},
		{
			"1 tag case",
			"host=localhost123",
			map[string]string{
				"host": "localhost123",
			},
		},
		{
			"empty case",
			"",
			map[string]string{},
		},
		{
			"garbage",
			"smth=smthelse=123",
			map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitTagsFlag(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitTagsFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}