package ram_image_generator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	pb "github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/proto_generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure" // Для упрощения не будем использовать SSL/TLS аутентификация
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	ImageGenerationTimeout = errors.New("image generation timeout")
	TooLongPromptError     = errors.New("too long prompt error")
	CensorshipError        = errors.New("user prompt or descriptions contains illegal content")
)

type imageUploadApiResponseImage struct {
	Url string `json:"url"`
}

type imageUploadApiResponse struct {
	StatusCode int                         `json:"status_code"`
	Image      imageUploadApiResponseImage `json:"image"`
}

func AuthInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func CreateGRPCConnection() pb.RamGeneratorClient {
	var conn *grpc.ClientConn
	var err error
	for {
		conn, err = grpc.NewClient(
			fmt.Sprintf("%s", config.Conf.GRPC.Hostname),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(AuthInterceptor(config.Conf.GRPC.Token)),
		)
		if err == nil {
			break
		}
		log.Print("retry gRPC connection, error: ", err)
		time.Sleep(2 * time.Second)
	}
	log.Print("GRPC подключен")
	return pb.NewRamGeneratorClient(conn)
}

func GenerateStartPrompt(context context.Context, grpcClient pb.RamGeneratorClient, userPrompt string) (string, error) {
	if len(userPrompt) > config.Conf.Generation.MaxPromptLen {
		return "", TooLongPromptError
	}
	prompt, err := grpcClient.GenerateStartPrompt(context, &pb.GenerateStartPromptRequest{UserPrompt: userPrompt})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			if st.String() == "User prompt contains illegal content" {
				return "", CensorshipError
			}
		}
		return "", err
	}
	return prompt.Prompt, nil
}

func GenerateHybridPrompt(context context.Context, grpcClient pb.RamGeneratorClient, userPrompt string, ramsDescription []string) (string, error) {
	if len(userPrompt) > config.Conf.Generation.MaxPromptLen {
		return "", TooLongPromptError
	}

	prompt, err := grpcClient.GenerateHybridPrompt(context, &pb.GenerateHybridPromptRequest{UserPrompt: userPrompt, RamDescriptions: ramsDescription})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			if st.String() == "User prompt or descriptions contains illegal content" {
				return "", CensorshipError
			}
		}
		return "", err
	}
	return prompt.Prompt, nil
}

func GenerateRamImage(context context.Context, grpcClient pb.RamGeneratorClient, prompt string) (string, error) {
	generatedImage, err := grpcClient.GenerateImage(context, &pb.GenerateImageRequest{Prompt: prompt, Style: config.Conf.Image.DefaultKandinskyStyle})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.DeadlineExceeded {
			return "", ImageGenerationTimeout
		}
		return "", err
	}
	return generatedImage.Image, nil
}

func UploadImage(base64Image string) (string, error) {
	fromData := url.Values{
		"key":    {config.Conf.AnotherTokens.FreeImageHostApiKey},
		"source": {base64Image},
	}
	resp, err := http.PostForm(fmt.Sprintf("https://freeimage.host/api/1/upload"), fromData)

	if err != nil {
		log.Println(err)
		return "", err
	}
	var jsonResp imageUploadApiResponse
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, &jsonResp)
	if err != nil {
		return "", err
	}
	if jsonResp.StatusCode != 200 {
		log.Println(string(data))
		return "", fmt.Errorf("unexpected image upload api error")
	}
	return jsonResp.Image.Url, nil
}

func GenerateDescription(context context.Context, grpcClient pb.RamGeneratorClient, url string) (string, error) {
	description, err := grpcClient.GenerateDescription(context, &pb.RamImageUrl{Url: url})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			return "", CensorshipError
		}
		return "", err
	}
	return description.Description, nil
}
