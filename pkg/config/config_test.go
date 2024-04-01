package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// testing (types are not public (not exposed outside)
type dummyInfo struct {
	Email string
	Array []int
}
type nestedInfo struct {
	NestedId    int
	NestedValue string
}
type keyInfo struct {
	Foo       string
	Seriously string
	Number    int
	Nested    *nestedInfo
}
type testConfig struct {
	Key       *keyInfo
	Dummy     *dummyInfo
	Dangerous bool
}

func loadConfig(t *testing.T, path string) *testConfig {
	conf := &testConfig{}
	err := LoadConfig(path, conf)
	if err != nil {
		t.Error(err)
	}
	return conf
}

func Test_Dev(t *testing.T) {

	_ = os.Setenv("ENV", "")

	c := loadConfig(t, "./fixtures")

	require.Equal(t, "bar", c.Key.Foo, "Key.Foo")
	require.Equal(t, 5, c.Key.Number, "Key.Number")

	require.Equal(t, 20, c.Key.Nested.NestedId, "NestedId")
	require.Equal(t, "test", c.Key.Nested.NestedValue, "NestedValue")

	//prod only keys
	require.Equal(t, "", c.Key.Seriously, "Seriously")
	require.Equal(t, false, c.Dangerous, "Dangerous")

	require.Equal(t, &dummyInfo{
		Email: "test@example.com",
		Array: []int{10, 20},
	}, c.Dummy)

}

func Test_Prod(t *testing.T) {

	_ = os.Setenv("ENV", "prod")

	c := loadConfig(t, "./fixtures")

	require.Equal(t, "Prod", c.Key.Seriously, "Seriously")
	require.Equal(t, true, c.Dangerous, "Dangerous")

}

func Test_LoadConfigFromEnv(t *testing.T) {
	//binding

	//add binding
	viper.SetDefault("Env", "dev")

	_ = viper.BindEnv("Env", "ENV")

	// bindings (Viper key to a ENV variable)
	_ = viper.BindEnv("Key.Foo", "KEY_FOO")
	_ = viper.BindEnv("Key.Number", "KEY_NUMBER")

	_ = viper.BindEnv("Key.Nested.NestedId", "KEY_NESTED_ID")
	_ = viper.BindEnv("Key.Nested.NestedValue", "KEY_NESTED_VALUE")

	_ = viper.BindEnv("Dummy.Email", "KEY_DUMMY_EMAIL")
	_ = viper.BindEnv("Dummy.Array", "KEY_DUMMY_ARRAY")

	//dev only
	_ = viper.BindEnv("Key.FooDev", "KEY_FOO_DEV")
	_ = viper.BindEnv("Key3.Hello", "KEY3_HELLO")

	//prod only
	_ = viper.BindEnv("Key.Seriously", "KEY_SERIOUSLY")
	_ = viper.BindEnv("Dangerous", "DANGEROUS")

	//set env vars

	vars := map[string]string{
		"KEY_FOO":          "env_foo",
		"KEY_NUMBER":       "10",
		"KEY_NESTED_ID":    "100",
		"KEY_NESTED_VALUE": "env_nested_value",
		"KEY_DUMMY_EMAIL":  "env_email",
		"KEY_DUMMY_ARRAY":  "1,2,3",

		"KEY_FOO_DEV": "env_key_foo_dev",
		"KEY3_HELLO":  "env_key3_hello",
	}

	_ = os.Setenv("ENV", "dev")

	keys := setEnvVars(t, vars)
	defer unsetEnvVars(t, keys)

	// defaults => _{env} => envVars
	c := loadConfig(t, "./fixtures")

	//test config vals - dev
	require.Equal(t, "env_foo", c.Key.Foo, "Key.Foo")
	require.Equal(t, 10, c.Key.Number, "Key.Number")

	require.Equal(t, 100, c.Key.Nested.NestedId, "NestedId")
	require.Equal(t, "env_nested_value", c.Key.Nested.NestedValue, "NestedValue")

	//prod only keys
	require.Equal(t, "", c.Key.Seriously, "Seriously")
	require.Equal(t, false, c.Dangerous, "Dangerous")

	require.Equal(t, &dummyInfo{
		Email: "env_email",
		Array: []int{1, 2, 3},
	}, c.Dummy)

	//test config vals - prod
	os.Setenv("ENV", "prod")

	//override env vars
	setEnvVars(t, map[string]string{
		"KEY_SERIOUSLY": "env_seriously",
		"DANGEROUS":     "true",
	})

	//reload again
	c = loadConfig(t, "./fixtures")

	require.Equal(t, "env_seriously", c.Key.Seriously, "Seriously")
	require.Equal(t, true, c.Dangerous, "Dangerous")

}

// // setEnvVars will set a map of key/value to environment variables
// // and also return a set of keys to be unset later
func setEnvVars(t *testing.T, vars map[string]string) []string {
	var keys []string
	for key, value := range vars {
		require.Nil(t, os.Setenv(key, value))
		keys = append(keys, key)
	}
	return keys
}

func unsetEnvVars(t *testing.T, keys []string) {
	for _, key := range keys {
		require.Nil(t, os.Unsetenv(key))
	}
}
