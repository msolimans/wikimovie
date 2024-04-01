package testutil

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/google/go-cmp/cmp"
)

func StripNewlines(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return strings.Replace(s, " ", "", -1)
}

// NowUTCTruncated will return now time and truncate millisecond since mongo does that automatically
func NowUTCTruncated() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func CreateQueryString(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	return values.Encode()
}

// ListContainsItem will return a bool if list contains the item determined by a cmp.Diff
func ListContainsItem(list interface{}, item interface{}) bool {
	s := reflect.ValueOf(list)
	for i := 0; i < s.Len(); i++ {
		if diff := cmp.Diff(s.Index(i).Interface(), item); diff == "" {
			return true
		}
	}
	return false
}

// LoadTestFixture will load a file from testutil/fixtures folder and unmarshal into the field
func LoadTestFixture(t *testing.T, fileName string, field interface{}) {
	_, b, _, ok := runtime.Caller(0)
	require.True(t, ok, "can not determine runtime caller")
	fixturePath := filepath.Join(filepath.Dir(b), "../../pkg/testutil/fixtures")

	file, err := ioutil.ReadFile(filepath.Join(fixturePath, fileName))
	require.NoError(t, err, "error reading file")

	err = json.Unmarshal(file, field)
	require.NoError(t, err, "error during unmarshal")
}

// GetTimeFromStr will parse a str as RFC3339 and return the time
func GetTimeFromStr(t *testing.T, str string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, str)
	require.Nil(t, err)
	return parsedTime
}
