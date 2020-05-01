package config

import (
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMain(t *testing.M) {
	err := os.Setenv(EnvVar, strings.Join([]string{
		filepath.Join("testdata", "1"),
		filepath.Join("testdata", "2"),
	}, string(os.PathListSeparator)))
	if err != nil {
		panic(err)
	}

	os.Exit(t.Run())
}

type testInput struct {
	name string
	args args
	want []byte
}

type args struct {
	n string
}

func testInputs(t *testing.T) chan *testInput {
	t.Helper()

	result := make(chan *testInput)
	go func() {
		defer close(result)
		result <- &testInput{
			"1/bytes",
			args{"bytes"},
			[]byte("1234567890"),
		}
		result <- &testInput{
			"2/string",
			args{"string"},
			[]byte("Hello, World! ✌"),
		}
		result <- &testInput{
			"1/user.json",
			args{"user.json"},
			[]byte(`{"username": "user"}`),
		}
		result <- &testInput{
			"2/userpass.json",
			args{"userpass.json"},
			[]byte(`{"username": "user", "password":  "pass"}`),
		}
	}()
	return result
}

func TestBytes(t *testing.T) {
	for tt := range testInputs(t) {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Bytes(tt.args.n)
			if err != nil {
				t.Errorf("Bytes() error = %v, wantErr %v", err, false)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	for tt := range testInputs(t) {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String(tt.args.n)
			if err != nil {
				t.Errorf("String() error = %v, wantErr %v", err, false)
				return
			}
			if want := string(tt.want); got != want {
				t.Errorf("String() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserinfo(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name    string
		args    args
		want    *url.Userinfo
		wantErr bool
	}{
		{
			"1/bytes",
			args{"bytes"},
			nil,
			true,
		},
		{
			"2/string",
			args{"string"},
			nil,
			true,
		},
		{
			"1/user.json",
			args{"user.json"},
			url.User("user"),
			false,
		},
		{
			"2/userpass.json",
			args{"userpass.json"},
			url.UserPassword("user", "pass"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Userinfo(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Userinfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Userinfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUrl(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name    string
		args    args
		want    *url.URL
		wantErr bool
	}{

		{
			"1/bytes",
			args{"bytes"},
			&url.URL{Path: "1234567890"},
			false,
		},
		{
			"2/string",
			args{"string"},
			&url.URL{Path: "Hello, World! ✌"},
			false,
		},
		{
			"1/user.json",
			args{"user.json"},
			nil,
			true,
		},
		{
			"2/userpass.json",
			args{"userpass.json"},
			nil,
			true,
		},
		{
			"1/http_url",
			args{"http_url"},
			&url.URL{
				Scheme:   "http",
				Host:     "google.com",
				RawQuery: "q=tuukka",
			},
			false,
		},
		{
			"2/db_url",
			args{"db_url"},
			&url.URL{
				Scheme:   "postgres",
				User:     url.UserPassword("user", "pass"),
				Host:     "localhost:5432",
				Path:     "/db",
				RawQuery: "sslmode=require",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Url(tt.args.n)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Url() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if !reflect.DeepEqual(got.String(), tt.want.String()) {
				t.Errorf("Url() got = %#v, want %#v", got.String(), tt.want.String())
			}
		})
	}
}
