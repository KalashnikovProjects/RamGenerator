package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/api"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	pb "github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/proto_generated"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TODO ПОЧИНИТЬ ТЕСТЫ

type expected struct {
	statusCode    int
	response      string
	responseRegex *regexp.Regexp
}

var testingHostname = "localhost"
var testingPort = 8083
var testingHost = fmt.Sprintf("%s:%d", testingHostname, testingPort)
var jwtRegexp = regexp.MustCompile("^[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+$")

type gRPCStub struct{}

func (g *gRPCStub) GenerateStartPrompt(ctx context.Context, in *pb.GenerateStartPromptRequest, opts ...grpc.CallOption) (*pb.RamImagePrompt, error) {
	return &pb.RamImagePrompt{Prompt: "happy ram"}, nil
}

func (g *gRPCStub) GenerateHybridPrompt(ctx context.Context, in *pb.GenerateHybridPromptRequest, opts ...grpc.CallOption) (*pb.RamImagePrompt, error) {
	return &pb.RamImagePrompt{Prompt: "very happy ram"}, nil
}

func (g *gRPCStub) GenerateImage(ctx context.Context, in *pb.GenerateImageRequest, opts ...grpc.CallOption) (*pb.RamImage, error) {
	file, err := os.Open("./test_ram_base64_image.txt")
	if err != nil {
		panic(fmt.Sprintf("cant open required for testing file test_ram_base64_image.txt, err: %v", err))
	}
	data, err := io.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("cant read required for testing file test_ram_base64_image.txt, err: %v", err))
	}
	return &pb.RamImage{Image: string(data)}, nil
}

func (g *gRPCStub) GenerateDescription(ctx context.Context, in *pb.RamImageUrl, opts ...grpc.CallOption) (*pb.RamDescription, error) {
	return &pb.RamDescription{Description: "wow very nice ram"}, nil
}

func RunPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres"),
		postgres.WithDatabase("ram_generator_test"),
		postgres.WithUsername("username"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(20*time.Second)),
	)
	return pgContainer, err
}

func setup() *postgres.PostgresContainer {
	var err error
	ctx := context.Background()
	pgContainer, err := RunPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("error running postgres container: %v", err)
	}

	port, _ := pgContainer.MappedPort(ctx, "5432/tcp")
	host, _ := pgContainer.Host(ctx)
	os.Setenv("ROOT_PATH", "../../..")
	os.Setenv("POSTGRES_HOST", fmt.Sprintf("%s:%s", host, port.Port()))
	os.Setenv("POSTGRES_DB", "ram_generator_test")
	os.Setenv("POSTGRES_USER", "username")
	os.Setenv("POSTGRES_PASSWORD", "password")
	os.Setenv("HMAC", "testHmac123")
	if godotenv.Load(".env") != nil {
		log.Fatal(".env file not found, it must contain FREE_IMAGE_HOST_API_KEY")
	}
	config.InitConfigs()
	log.Println("starting api server...")

	pgConnectionString := database.GeneratePostgresConnectionString(config.Conf.Database.User, config.Conf.Database.Password, config.Conf.Database.Host, config.Conf.Database.DBName)
	db := database.CreateDBConnectionContext(ctx, pgConnectionString)

	server := api.NewRamGeneratorServer(ctx, fmt.Sprintf(":%d", testingPort), db, &gRPCStub{})
	go func(server *http.Server) {
		err := api.ServeServer(ctx, server)
		if err != nil {
			log.Fatal("server shutdown with error:", err)
		}
	}(server)

	time.Sleep(3 * time.Second)
	return pgContainer
}

func teardown(pgContainer *postgres.PostgresContainer) {
	if pgContainer != nil {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("error terminate postgres container: %v", err)
		}
	}
}

func registerUser(username string, password string) (string, error) {
	inputReader := strings.NewReader(fmt.Sprintf(`{"username": "%s", "password":"%s"}`, username, password))
	resp, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/register", testingHost), inputReader)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	client := &http.Client{}
	res, err := client.Do(resp)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()
	text, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed register request, error: %d %s", res.StatusCode, strings.TrimSpace(string(text)))
	}

	resp, err = http.NewRequest("GET", fmt.Sprintf("http://%s/api/users/%s", testingHost, username), strings.NewReader(""))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	res, err = client.Do(resp)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to register user, get user returns: %d %s", res.StatusCode, string(body))
	}
	return string(text), nil
}

func deleteUser(username string, token string) error {
	resp, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s/api/users/%s", testingHost, username), strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	resp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{}
	res, err := client.Do(resp)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		errorText, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %v", err)
		}
		return fmt.Errorf("failed delete request, error: %d %s", res.StatusCode, strings.TrimSpace(string(errorText)))
	}

	resp, err = http.NewRequest("GET", fmt.Sprintf("http://%s/api/users/%s", testingHost, username), strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	resp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err = client.Do(resp)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete user, get user return: %d %s", res.StatusCode, string(body))
	}
	return nil
}

func TestMain(m *testing.M) {
	pgContainer := setup()
	code := m.Run()
	teardown(pgContainer)
	os.Exit(code)
}

func TestLoginRegister(t *testing.T) {
	var token string
	defer func() {
		err := deleteUser("test1_name", token)
		if err != nil {
			t.Error(err)
		}
	}()
	testCases := []struct {
		name     string
		method   string
		url      string
		body     string
		expected expected
	}{
		{
			name:   "normal register",
			method: "POST",
			url:    "api/register",
			body:   `{"username": "test1_name", "password": "1234"}`,
			expected: expected{
				statusCode:    http.StatusOK,
				responseRegex: jwtRegexp,
			},
		},
		{
			name:   "bad json",
			method: "POST",
			url:    "api/register",
			body:   `{"usern`,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   "json decoding error",
			},
		},
		{
			name:   "too big username",
			method: "POST",
			url:    "api/register",
			body:   `{"username": "alooooooooooooooooooooooooooooooooooooooooooooooooooaloooooooooooooooooooooooooooooooooooooooooooooooooo", "password": "12345"}`,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("username must be 3-%d lenght and contain only English letters, numbers and _", config.Conf.Users.MaxUsernameLen),
			},
		},
		{
			name:   "bad username",
			method: "POST",
			url:    "api/register",
			body:   `{"username": "user name", "password": "12345"}`,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("username must be 3-%d lenght and contain only English letters, numbers and _", config.Conf.Users.MaxUsernameLen),
			},
		},
		{
			name:   "name already taken",
			method: "POST",
			url:    "api/register",
			body:   `{"username": "test1_name", "password": "qwerty"}`,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("username %s is already taken", "test1_name"),
			},
		},
		{
			name:   "required fields are not specified",
			method: "POST",
			url:    "api/register",
			body:   `{"username": "testfawfa2w"}`,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   "required fields are not specified: password",
			},
		},
		{
			name:   "login with correct credentials",
			method: "POST",
			url:    "api/login",
			body:   `{"username": "test1_name", "password": "1234"}`,
			expected: expected{
				statusCode:    http.StatusOK,
				responseRegex: jwtRegexp,
			},
		},
		{
			name:   "login with incorrect name",
			method: "POST",
			url:    "api/login",
			body:   `{"username": "my_nafme", "password": "12345678"}`,
			expected: expected{
				statusCode: http.StatusNotFound,
				response:   "no users with username = my_nafme",
			},
		},
		{
			name:   "login with incorrect password",
			method: "POST",
			url:    "api/login",
			body:   `{"username": "test1_name", "password": "12345f6awfawf78"}`,
			expected: expected{
				statusCode: http.StatusUnauthorized,
				response:   "wrong password",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			resp, err := http.NewRequest(tc.method, fmt.Sprintf("http://%s/%s", testingHost, tc.url), strings.NewReader(tc.body))
			if err != nil {
				t.Errorf("failed to create request: %v", err)
				return
			}
			resp.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			res, err := client.Do(resp)
			if err != nil {
				t.Errorf("failed to send request: %v", err)
				return
			}
			defer res.Body.Close()
			textBytes, err := io.ReadAll(res.Body)
			text := strings.TrimSpace(string(textBytes))
			if err != nil {
				t.Errorf("failed to read body: %v", err)
				return
			}
			if res.StatusCode != tc.expected.statusCode {
				t.Errorf("unexpected status code: got %d, want %d, body: %s", res.StatusCode, tc.expected.statusCode, text)
				return
			}
			if tc.expected.response != "" && text != tc.expected.response {
				t.Errorf("wrong response: got %s, want %s", text, tc.expected.response)
				return
			}
			if tc.expected.responseRegex != nil && !tc.expected.responseRegex.Match([]byte(text)) {
				t.Errorf("response dont match expected response regex: got %s, want %s", text, tc.expected.responseRegex.String())
				return
			}
			if tc.name == "login with correct credentials" {
				token = text
			}
		})
	}
}

func TestAuthPermissions(t *testing.T) {
	reqBody := `{"username": "test2_name", "password": "1234"}`
	resp, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/register", testingHost), strings.NewReader(reqBody))
	if err != nil {
		t.Errorf("failed to create request: %v", err)
		return
	}
	resp.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(resp)
	if err != nil {
		t.Errorf("failed to send request: %v", err)
		return
	}
	defer res.Body.Close()
	tokenBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
		return
	}
	registerToken := strings.TrimSpace(string(tokenBytes))
	defer func() {
		err := deleteUser("test2_name", registerToken)
		if err != nil {
			t.Error(err)
		}
	}()

	reqBody = `{"username": "test2_name", "password": "1234"}`
	resp, err = http.NewRequest("POST", fmt.Sprintf("http://%s/api/login", testingHost), strings.NewReader(reqBody))
	if err != nil {
		t.Errorf("failed to create request: %v", err)
		return
	}
	resp.Header.Set("Content-Type", "application/json")
	client = &http.Client{}
	res, err = client.Do(resp)
	if err != nil {
		t.Errorf("failed to send request: %v", err)
		return
	}
	defer res.Body.Close()
	tokenBytes, err = io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
		return
	}
	loginToken := strings.TrimSpace(string(tokenBytes))

	testCases := []struct {
		name      string
		method    string
		url       string
		needToken bool
	}{
		{
			name:      "register, not need token",
			method:    "POST",
			url:       "api/register",
			needToken: false,
		},
		{
			name:      "login, not need token",
			method:    "POST",
			url:       "api/login",
			needToken: false,
		},
		{
			name:      "get user, not need token",
			method:    "GET",
			url:       "api/users/123",
			needToken: false,
		},
		{
			name:      "put user, need token",
			method:    "PUT",
			url:       "api/users/123",
			needToken: true,
		},
		{
			name:      "patch user, need token",
			method:    "PATCH",
			url:       "api/users/123",
			needToken: true,
		},
		{
			name:      "delete user, need token",
			method:    "DELETE",
			url:       "api/users/123",
			needToken: true,
		},
		{
			name:      "get rams, not need token",
			method:    "GET",
			url:       "api/users/123/rams",
			needToken: false,
		},
		{
			name:      "get ram, not need token",
			method:    "GET",
			url:       "api/users/123/rams/1234",
			needToken: false,
		},
		//{
		//	name:      "put ram, need token",
		//	method:    "PUT",
		//	url:       "api/users/123/rams/1234",
		//	needToken: true,
		//},
		//{
		//	name:      "patch ram, need token",
		//	method:    "PATCH",
		//	url:       "api/users/123/rams/1234",
		//	needToken: true,
		//},
		//{
		//	name:      "delete ram, need token",
		//	method:    "DELETE",
		//	url:       "api/users/123/rams/1234",
		//	needToken: true,
		//},
	}

	for _, tc := range testCases {
		for tokenNum, testToken := range []string{"", "bad token", registerToken, loginToken} {
			name := tc.name
			niceToken := false
			switch tokenNum {
			case 0:
				name += " (no token)"
			case 1:
				name += " (bad token)"
			case 2:
				niceToken = true
				name += " (register token)"
			case 3:
				niceToken = true
				name += " (login token)"
			}
			t.Run(name, func(t *testing.T) {
				resp, err := http.NewRequest(tc.method, fmt.Sprintf("http://%s/%s", testingHost, tc.url), strings.NewReader(""))
				if err != nil {
					t.Errorf("failed to create request: %v", err)
					return
				}
				resp.Header.Set("Content-Type", "application/json")
				if testToken != "" {
					resp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))
				}

				client := &http.Client{}
				res, err := client.Do(resp)
				if err != nil {
					t.Errorf("failed to send request: %v", err)
					return
				}
				defer res.Body.Close()
				if tc.needToken && !niceToken && res.StatusCode != http.StatusUnauthorized {
					t.Errorf("unexpected status code: got %d, want 401 authorization error", res.StatusCode)
					return
				}
				if (!tc.needToken || niceToken) && res.StatusCode == http.StatusUnauthorized {
					t.Errorf("unexpected 401 authorization error, want another status code")
					return
				}
			})
		}
	}
}

func TestUsersResource(t *testing.T) {
	var token, anotherToken string
	token, err := registerUser("test3_name", "1234")
	if err != nil {
		t.Error(err)
	}
	anotherToken, err = registerUser("test3_another_name", "1234")
	if err != nil {
		t.Error(err)
	}

	testCases := []struct {
		name     string
		method   string
		url      string
		body     string
		token    string
		expected expected
	}{
		{
			name:   "404 bad url",
			method: "POST",
			url:    "api/users4/test3_name",
			body:   ``,
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:   "405 bad method",
			method: "POST",
			url:    "api/users/test3_name",
			body:   ``,
			expected: expected{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "get unknown user",
			method: "GET",
			url:    "api/users/fake_user",
			body:   ``,
			expected: expected{
				statusCode: http.StatusNotFound,
				response:   "no users with username = fake_user",
			},
		},
		{
			name:   "get user",
			method: "GET",
			url:    "api/users/test3_name",
			body:   ``,
			expected: expected{
				statusCode:    http.StatusOK,
				responseRegex: regexp.MustCompile(`\{"id":4,"username":"test3_name","daily_ram_generation_time":0,"rams_generated_last_day":0,"avatar_ram_id":0,"avatar_box":\[\[(\d+(?:\.\d+)?),(\d+(?:\.\d+)?)],\[(\d+(?:\.\d+)?),(\d+(?:\.\d+)?)]]}`),
			},
		},
		{
			name:   "patch user",
			method: "PATCH",
			url:    "api/users/test3_name",
			body:   `{"username":"test3_name_edited","password":"qwerty","avatar_box": [[1,1],[2, 2]], "last_ram_generated":52}`,
			token:  token,
			expected: expected{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "patch user, username already taken",
			method: "PATCH",
			url:    "api/users/test3_name_edited",
			body:   `{"username":"test3_another_name"}`,
			token:  token,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("username test3_another_name is already taken"),
			},
		},
		{
			name:   "patch user, bad username",
			method: "PATCH",
			url:    "api/users/test3_name_edited",
			body:   `{"username":"haha error"}`,
			token:  token,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("username must be 3-%d lenght and contain only English letters, numbers and _", config.Conf.Users.MaxUsernameLen),
			},
		},
		{
			name:   "put user, all fields must be filled",
			method: "PUT",
			url:    "api/users/test3_name_edited",
			body:   `{"username": "awfawfawfa"}`,
			token:  token,
			expected: expected{
				statusCode: http.StatusBadRequest,
				response:   "all fields must be filled for PUT request",
			},
		},
		{
			name:   "get user after put",
			method: "GET",
			url:    "api/users/test3_name_edited",
			body:   ``,
			expected: expected{
				statusCode:    http.StatusOK,
				responseRegex: regexp.MustCompile(`\{"id":4,"username":"test3_name_edited","daily_ram_generation_time":0,"rams_generated_last_day":0,"avatar_ram_id":0,"avatar_box":\[\[(\d+(?:\.\d+)?),(\d+(?:\.\d+)?)],\[(\d+(?:\.\d+)?),(\d+(?:\.\d+)?)]]}`),
			},
		},
		{
			name:   "get user, old username",
			method: "GET",
			url:    "api/users/test3_name",
			body:   ``,
			expected: expected{
				statusCode: http.StatusNotFound,
				response:   "no users with username = test3_name",
			},
		},
		{
			name:   "patch another user, forbidden",
			method: "PATCH",
			url:    "api/users/test3_another_name",
			body:   `{"password":"qwerty1"}`,
			token:  token,
			expected: expected{
				statusCode: http.StatusForbidden,
				response:   "you can't edit another user",
			},
		},
		{
			name:   "delete fake user",
			method: "DELETE",
			url:    "api/users/fake_user",
			body:   ``,
			token:  token,
			expected: expected{
				statusCode: http.StatusForbidden,
				response:   "you can't delete another user",
			},
		},
		{
			name:   "delete another user, forbidden",
			method: "DELETE",
			url:    "api/users/test3_another_name",
			body:   ``,
			token:  token,
			expected: expected{
				statusCode: http.StatusForbidden,
				response:   "you can't delete another user",
			},
		},
		{
			name:   "get another user after try to delete",
			method: "GET",
			url:    "api/users/test3_another_name",
			body:   ``,
			expected: expected{
				statusCode:    http.StatusOK,
				responseRegex: regexp.MustCompile(`\{"id":5,"username":"test3_another_name","daily_ram_generation_time":0,"rams_generated_last_day":0,"avatar_ram_id":0,"avatar_box":\[\[(\d+(?:\.\d+)?),(\d+(?:\.\d+)?)],\[(\d+(?:\.\d+)?),(\d+(?:\.\d+)?)]]}`),
			},
		},
		{
			name:   "delete user",
			method: "DELETE",
			url:    "api/users/test3_name_edited",
			body:   ``,
			token:  token,
			expected: expected{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "get user after delete",
			method: "GET",
			url:    "api/users/test3_name_edited",
			body:   ``,
			expected: expected{
				statusCode: http.StatusNotFound,
				response:   "no users with username = test3_name_edited",
			},
		},
		{
			name:   "delete another user, ok",
			method: "DELETE",
			url:    "api/users/test3_another_name",
			body:   ``,
			token:  anotherToken,
			expected: expected{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "get another user after delete",
			method: "GET",
			url:    "api/users/test3_another_name",
			body:   ``,
			expected: expected{
				statusCode: http.StatusNotFound,
				response:   "no users with username = test3_another_name",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.NewRequest(tc.method, fmt.Sprintf("http://%s/%s", testingHost, tc.url), strings.NewReader(tc.body))
			if err != nil {
				t.Errorf("failed to create request: %v", err)
				return
			}
			resp.Header.Set("Content-Type", "application/json")
			if tc.token != "" {
				resp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))
			}
			client := &http.Client{}
			res, err := client.Do(resp)
			if err != nil {
				t.Errorf("failed to send request: %v", err)
				return
			}
			defer res.Body.Close()
			textBytes, err := io.ReadAll(res.Body)
			text := strings.TrimSpace(string(textBytes))
			if err != nil {
				t.Errorf("failed to read body: %v", err)
				return
			}
			if res.StatusCode != tc.expected.statusCode {
				t.Errorf("unexpected status code: got %d, want %d, body: %s", res.StatusCode, tc.expected.statusCode, text)
				return
			}
			if tc.expected.response != "" && text != tc.expected.response {
				t.Errorf("wrong response: got %s, want %s", text, tc.expected.response)
				return
			}
			if tc.expected.responseRegex != nil && !tc.expected.responseRegex.Match([]byte(text)) {
				t.Errorf("response dont match expected response regex: got %s, want %s", text, tc.expected.responseRegex.String())
				return
			}
		})
	}
}

// TODO: Добавить тесты с test_rams_another_user

func TestRamsResource(t *testing.T) {
	var token string
	token, err := registerUser("test_rams_user", "password123")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := deleteUser("test_rams_user", token)
		if err != nil {
			t.Error(err)
		}
	}()
	anotherToken, err := registerUser("test_rams_another_user", "password123")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := deleteUser("test_rams_another_user", anotherToken)
		if err != nil {
			t.Error(err)
		}
	}()

	t.Run("GetRams", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/api/users/test_rams_user/rams", testingHost))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			text, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			t.Fatalf("Expected status OK, got: %d %s", resp.StatusCode, text)
		}

		var rams []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&rams)
		if err != nil {
			t.Fatal(err)
		}

		if len(rams) != 0 {
			t.Fatalf("Expected empty array of rams, got %v", rams)
		}
	})

	t.Run("GetRams bad username", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/api/users/ahahahahah/rams", testingHost))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			text, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			t.Fatalf("Expected status not found, got: %d %s", resp.StatusCode, text)
		}
	})

	t.Run("CreateRam", func(t *testing.T) {
		ctx := context.Background()

		url := fmt.Sprintf("ws://%s/api/users/test_rams_user/ws/generate-ram", testingHost)

		ws, resp, err := websocket.DefaultDialer.Dial(url, http.Header{})
		if err != nil {
			if errors.Is(err, websocket.ErrBadHandshake) {
				text, _ := io.ReadAll(resp.Body)
				defer resp.Body.Close()
				t.Fatalf("Error connection ws/generate-ram: %d %s", resp.StatusCode, text)
			}
			t.Fatal(err)
		}
		defer ws.Close()
		err = ws.WriteMessage(websocket.TextMessage, []byte(token))
		if err != nil {
			t.Fatal(err)
		}

		_, message, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}

		if strings.TrimSpace(string(message)) != `{"status":"need first ram prompt"}` {
			t.Fatalf(`Expected '{"status":"need first ram prompt"}', got '%s'`, strings.TrimSpace(string(message)))
		}

		err = ws.WriteMessage(websocket.TextMessage, []byte("Generate a cool ram"))
		if err != nil {
			t.Fatal(err)
		}

		_, message, err = ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(string(message)) != `{"clicks":100,"status":"need clicks"}` {
			t.Fatalf(`Expected '{"clicks":100,"status":"need clicks"}', got '%s'`, string(message))
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second*15)
		defer cancel()
		go func() {
			for _ = range 14 {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(1 * time.Second)
					err = ws.WriteMessage(websocket.TextMessage, []byte("15"))
					if err != nil {
						t.Error(err)
						return
					}
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err = ws.ReadMessage()
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println(string(message))

				var response map[string]any
				err = json.Unmarshal(message, &response)
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := response["id"]; ok {
					_, okImageUrl := response["image_url"]
					_, okDescription := response["description"]
					if !okImageUrl || !okDescription {
						t.Fatalf("Incomplete response: %v", response)
					}
					return
				} else {
					switch response["status"] {
					case "image generated":
					case "success clicked":
						continue
					}
				}
			}

		}
	})
	t.Run("CreateRam rate limit", func(t *testing.T) {
		url := fmt.Sprintf("ws://%s/api/users/test_rams_user/ws/generate-ram", testingHost)

		ws, resp, err := websocket.DefaultDialer.Dial(url, http.Header{})
		if err != nil {
			if errors.Is(err, websocket.ErrBadHandshake) {
				text, _ := io.ReadAll(resp.Body)
				defer resp.Body.Close()
				t.Fatalf("Error connection ws/generate-ram: %d %s", resp.StatusCode, text)
			}
			t.Fatal(err)
		}
		defer ws.Close()
		err = ws.WriteMessage(websocket.TextMessage, []byte(token))
		if err != nil {
			t.Fatal(err)
		}

		_, message, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		var response map[string]any
		err = json.Unmarshal(message, &response)
		if err != nil {
			t.Fatal(err)
		}

		if strings.HasPrefix(response["error"].(string), "you can generate next ram in") && response["code"] == 429 {
			t.Fatalf(`Expected error rate limit, got '%s'`, strings.TrimSpace(string(message)))
		}
	})
	t.Run("CreateRam another user", func(t *testing.T) {
		url := fmt.Sprintf("ws://%s/api/users/test_rams_user/ws/generate-ram", testingHost)

		ws, resp, err := websocket.DefaultDialer.Dial(url, http.Header{})
		if err != nil {
			if errors.Is(err, websocket.ErrBadHandshake) {
				text, _ := io.ReadAll(resp.Body)
				defer resp.Body.Close()
				t.Fatalf("Error connection ws/generate-ram: %d %s", resp.StatusCode, text)
			}
			t.Fatal(err)
		}
		defer ws.Close()
		err = ws.WriteMessage(websocket.TextMessage, []byte(token))
		if err != nil {
			t.Fatal(err)
		}

		_, message, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		var response map[string]any
		err = json.Unmarshal(message, &response)
		if err != nil {
			t.Fatal(err)
		}

		if strings.HasPrefix(response["error"].(string), "it's not your ram") && response["code"] == 403 {
			t.Fatalf(`Expected forbidden, got '%s'`, strings.TrimSpace(string(message)))
		}
	})
	t.Run("GetRam_GetRams", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/api/users/%s/rams/%d", testingHost, "test_rams_user", 1))
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			text, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status OK, got: %d %s", resp.StatusCode, text)
		}

		var ram map[string]any
		err = json.NewDecoder(resp.Body).Decode(&ram)
		if err != nil {
			t.Fatal(err)
		}

		if ram["id"] == 0 || ram["image_url"] == "" || ram["description"] == "" {
			t.Fatalf("Unexpected ram data: %v", ram)
		}
		resp, err = http.Get(fmt.Sprintf("http://%s/api/users/%s/rams", testingHost, "test_rams_user"))
		if resp.StatusCode != http.StatusOK {
			text, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			t.Fatalf("Expected status OK, got: %d %s", resp.StatusCode, text)
		}

		var rams []map[string]any
		err = json.NewDecoder(resp.Body).Decode(&rams)
		if err != nil {
			t.Fatal(err)
		}
		if !(rams[0]["id"] == ram["id"] && rams[0]["description"] == ram["description"] && rams[0]["image_url"] == ram["image_url"] && rams[0]["taps"] == ram["taps"] && rams[0]["user_id"] == ram["user_id"]) {
			t.Fatalf("Wrong response from get rams: %v, expected: %v", rams, []map[string]any{ram})
		}
	})

	t.Run("GetRam bad user and bad ram", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/api/users/ahahahahah/rams/1", testingHost))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			text, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			t.Fatalf("Expected status not found, got: %d %s", resp.StatusCode, text)
		}

		resp, err = http.Get(fmt.Sprintf("http://%s/api/users/test_rams_user/rams/1234", testingHost))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			text, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			t.Fatalf("Expected status not found, got: %d %s", resp.StatusCode, text)
		}
	})
}
