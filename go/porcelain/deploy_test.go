package porcelain

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	apiClient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/netlify/open-api/go/models"
	"github.com/netlify/open-api/go/plumbing/operations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLFSSha(t *testing.T) {
	t.Run("test with not a pointer file", func(t *testing.T) {
		file := strings.NewReader("Not a pointer file")
		data, err := readLFSData(file)
		if err != nil {
			t.Fatal(err)
		}

		if data != nil {
			t.Fatal("expected data to be nil without proper formatting")
		}
	})

	t.Run("test with v1 pointer", func(t *testing.T) {
		content := `version https://git-lfs.github.com/spec/v1
oid sha256:7e56e498ccb4cbb9c672e1aed6710fb91b2fd314394a666c11c33b2059ea3d71
size 1743570
`
		file := strings.NewReader(content)
		data, err := readLFSData(file)
		if err != nil {
			t.Fatal(err)
		}

		if data.SHA != "7e56e498ccb4cbb9c672e1aed6710fb91b2fd314394a666c11c33b2059ea3d71" {
			t.Fatalf("expected `7e56e498ccb4cbb9c672e1aed6710fb91b2fd314394a666c11c33b2059ea3d71`, got `%v`", data.SHA)
		}

		if data.Size != 1743570 {
			t.Fatalf("expected `1743570`, got `%v`", data.Size)
		}
	})
}

func TestAddWithLargeMedia(t *testing.T) {
	files := newDeployFiles()
	tests := []struct {
		rel string
		sum string
	}{
		{"foo.jpg", "sum1"},
		{"bar.jpg", "sum2"},
		{"baz.jpg", "sum3:originalsha"},
	}

	for _, test := range tests {
		file := &FileBundle{}
		file.Sum = test.sum
		files.Add(test.rel, file)
	}

	out := files.Hashed["sum3"]
	if len(out) != 1 {
		t.Fatalf("expected `%d`, got `%d`", 1, len(out))
	}
	out2 := files.Sums["baz.jpg"]
	if out2 != "sum3:originalsha" {
		t.Fatalf("expected `%v`, got `%v`", "sum3:originalsha", out2)
	}
}

func TestOpenAPIClientWithWeirdResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(408)
		rw.Write([]byte(`{ "foo": "bar", "message": "a message", "code": 408 }`))
	}))
	defer server.Close()

	httpClient := http.DefaultClient
	authInfo := runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
		r.SetHeaderParam("User-Agent", "buildbot")
		r.SetHeaderParam("Authorization", "Bearer 1234")
		return nil
	})

	hu, _ := url.Parse(server.URL)
	tr := apiClient.NewWithClient(hu.Host, "/api/v1", []string{"http"}, httpClient)
	client := NewRetryable(tr, strfmt.Default, 1)

	body := ioutil.NopCloser(bytes.NewReader([]byte("hello world")))
	params := operations.NewUploadDeployFileParams().WithDeployID("1234").WithPath("foo/bar/biz").WithFileBody(body)
	_, operationError := client.Operations.UploadDeployFile(params, authInfo)
	require.Error(t, operationError)
	require.Equal(t, "[PUT /deploys/{deploy_id}/files/{path}][408] uploadDeployFile default  &{Code:408 Message:a message}", operationError.Error())

}

func TestConcurrentFileUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(408)
		rw.Write([]byte(`{ "foo": "bar", "message": "a message", "code": 408 }`))
	}))
	defer server.Close()

	httpClient := http.DefaultClient
	authInfo := runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
		r.SetHeaderParam("User-Agent", "buildbot")
		r.SetHeaderParam("Authorization", "Bearer 1234")
		return nil
	})

	hu, _ := url.Parse(server.URL)
	tr := apiClient.NewWithClient(hu.Host, "/api/v1", []string{"http"}, httpClient)
	client := NewRetryable(tr, strfmt.Default, 1)
	for i := 0; i < 30; i++ {
		go func() {
			body := ioutil.NopCloser(bytes.NewReader([]byte("hello world")))
			params := operations.NewUploadDeployFileParams().WithDeployID("1234").WithPath("foo/bar/biz").WithFileBody(body)
			_, _ = client.Operations.UploadDeployFile(params, authInfo)
		}()
	}
}

func TestWalk_IgnoreNodeModulesInRoot(t *testing.T) {
	dir, err := ioutil.TempDir("", "deploy")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	err = os.Mkdir(filepath.Join(dir, "node_modules"), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, "node_modules", "root-package"), []byte{}, 0644)
	require.Nil(t, err)

	err = os.MkdirAll(filepath.Join(dir, "more", "node_modules"), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, "more", "node_modules", "inner-package"), []byte{}, 0644)
	require.Nil(t, err)

	files, err := walk(dir, mockObserver{}, false, false)
	require.Nil(t, err)
	assert.NotNil(t, files.Files["node_modules/root-package"])
	assert.NotNil(t, files.Files["more/node_modules/inner-package"])

	// When deploy directory == build directory, ignore node_modules in deploy directory root.
	files, err = walk(dir, mockObserver{}, false, true)
	require.Nil(t, err)
	assert.Nil(t, files.Files["node_modules/root-package"])
	assert.NotNil(t, files.Files["more/node_modules/inner-package"])
}

func TestReadZipRuntime(t *testing.T) {
	runtime, err := readZipRuntime("../internal/data/hello-rs-function-test.zip")
	if err != nil {
		t.Fatalf("unexpected error reading zip file: %v", err)
	}

	if runtime != "rs" {
		t.Fatalf("unexpected runtime value, expected='rs', got='%s'", runtime)
	}
}

type mockObserver struct{}

func (m mockObserver) OnSetupWalk() error                         { return nil }
func (m mockObserver) OnSuccessfulStep(*FileBundle) error         { return nil }
func (m mockObserver) OnSuccessfulWalk(*models.DeployFiles) error { return nil }
func (m mockObserver) OnFailedWalk()                              {}

func (m mockObserver) OnSetupDelta(*models.DeployFiles) error                      { return nil }
func (m mockObserver) OnSuccessfulDelta(*models.DeployFiles, *models.Deploy) error { return nil }
func (m mockObserver) OnFailedDelta(*models.DeployFiles)                           {}

func (m mockObserver) OnSetupUpload(*FileBundle) error      { return nil }
func (m mockObserver) OnSuccessfulUpload(*FileBundle) error { return nil }
func (m mockObserver) OnFailedUpload(*FileBundle)           {}
