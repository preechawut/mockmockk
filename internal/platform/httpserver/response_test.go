package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestStrictJSONDecoding(t *testing.T) {
	router := fiber.New(fiber.Config{ErrorHandler: Error, BodyLimit: 128 * 1024})
	router.Post("/", func(c fiber.Ctx) error {
		var target struct {
			Name string `json:"name"`
		}

		err := Decode(c, &target)
		if err != nil {
			return err
		}

		return JSON(c, http.StatusOK, target)
	})

	cases := []struct {
		name   string
		body   string
		status int
	}{
		{"unknown field", `{"unknown":1}`, http.StatusBadRequest},
		{"multiple values", `{"name":"a"} {"name":"b"}`, http.StatusBadRequest},
		{"malformed", `{"name":`, http.StatusBadRequest},
		{"oversized", strings.Repeat(" ", 65*1024) + `{"name":"a"}`, http.StatusBadRequest},
		{"valid", `{"name":"ข้าวผัด"}`, http.StatusOK},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")

			response, err := router.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			if response.StatusCode != test.status {
				t.Fatalf("expected %d got %d", test.status, response.StatusCode)
			}

			if test.status == http.StatusBadRequest {
				var payload struct {
					OK    *bool `json:"ok"`
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}

				err := json.NewDecoder(response.Body).Decode(&payload)
				if err != nil {
					t.Fatal(err)
				}

				if payload.OK == nil || *payload.OK || payload.Error.Code != "invalid" {
					t.Fatalf("unexpected error code: %s", payload.Error.Code)
				}
			}
		})
	}
}
