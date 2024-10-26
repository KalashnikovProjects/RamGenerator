package ram_generator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	pb "github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/proto_generated"
	"github.com/rivo/uniseg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

var (
	ImageGenerationTimeout     = errors.New("image generation timeout")
	ImageGenerationUnavailable = errors.New("image generation unavailable")
	TooLongPromptError         = errors.New("too long prompt error")
	CensorshipError            = errors.New("user prompt or descriptions contains illegal content")
	NoRamError                 = errors.New("no ram on final image")

	InternalPromptError      = errors.New("internal prompt generating error")
	InternalImageError       = errors.New("internal image generating error")
	InternalUploadError      = errors.New("internal image upload error")
	InternalDescriptionError = errors.New("internal description generating error")
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
	slog.Info("connecting gRPC")

	for {
		conn, err = grpc.NewClient(
			fmt.Sprintf("%s", config.Conf.GRPC.Host),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(AuthInterceptor(config.Conf.GRPC.Token)),
		)
		if err == nil {
			break
		}
		slog.Debug("retry gRPC connection", slog.String("error", err.Error()))
		time.Sleep(2 * time.Second)
	}
	slog.Info("gRPC connected")
	return pb.NewRamGeneratorClient(conn)
}

func GenerateStartPrompt(ctx context.Context, grpcClient pb.RamGeneratorClient, userPrompt string) (string, error) {
	if uniseg.GraphemeClusterCount(userPrompt) > config.Conf.Generation.MaxPromptLen {
		return "", TooLongPromptError
	}
	prompt, err := grpcClient.GenerateStartPrompt(ctx, &pb.GenerateStartPromptRequest{UserPrompt: userPrompt})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			if st.String() == "User prompt contains illegal content" {
				return "", CensorshipError
			}
		}
		slog.Error("generate start prompt grpc request error", slog.String("error", err.Error()), slog.String("status", st.Message()))
		return "", InternalPromptError
	}
	return prompt.Prompt, nil
}

func GenerateHybridPrompt(ctx context.Context, grpcClient pb.RamGeneratorClient, userPrompt string, ramsDescription []string) (string, error) {
	if uniseg.GraphemeClusterCount(userPrompt) > config.Conf.Generation.MaxPromptLen {
		return "", TooLongPromptError
	}

	prompt, err := grpcClient.GenerateHybridPrompt(ctx, &pb.GenerateHybridPromptRequest{UserPrompt: userPrompt, RamDescriptions: ramsDescription})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			if st.String() == "User prompt or descriptions contains illegal content" {
				return "", CensorshipError
			}
		}
		slog.Error("generate hybrid prompt grpc request error", slog.String("error", err.Error()), slog.String("status", st.Message()))
		return "", InternalPromptError
	}
	return prompt.Prompt, nil
}

func GenerateRamImage(ctx context.Context, grpcClient pb.RamGeneratorClient, prompt string) (string, error) {
	generatedImage, err := grpcClient.GenerateImage(ctx, &pb.GenerateImageRequest{Prompt: prompt, Style: config.Conf.Image.DefaultKandinskyStyle})
	if err != nil {
		st, ok := status.FromError(err)
		slog.Error("generate ram image grpc request error", slog.String("error", err.Error()), slog.String("status", st.Message()))
		if ok && st.Code() == codes.DeadlineExceeded {
			return "", ImageGenerationTimeout
		}
		if ok && st.Code() == codes.InvalidArgument {
			return "", CensorshipError
		}
		if ok && st.Code() == codes.Internal {
			if st.String() == "Image generation service unavailable" {
				return "", ImageGenerationUnavailable
			}
		}
		return "", InternalImageError
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
		slog.Error("image upload request error", slog.String("error", err.Error()))
		return "", InternalUploadError
	}
	var jsonResp imageUploadApiResponse
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("image upload request read error", slog.String("error", err.Error()))
		return "", InternalUploadError
	}
	err = json.Unmarshal(data, &jsonResp)
	if err != nil {
		slog.Error("image upload unmarshal json error", slog.String("error", err.Error()))
		return "", InternalUploadError
	}
	if jsonResp.StatusCode != 200 {
		slog.Error("image upload request error", slog.Int("statusCode", jsonResp.StatusCode), slog.Any("response", jsonResp))
		return "", InternalUploadError
	}
	return jsonResp.Image.Url, nil
}

func GenerateDescription(ctx context.Context, grpcClient pb.RamGeneratorClient, url string) (string, error) {
	description, err := grpcClient.GenerateDescription(ctx, &pb.RamImageUrl{Url: url})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			if st.String() == "Image does not contain ram" {
				return "", NoRamError
			} else {
				return "", CensorshipError
			}
		}
		slog.Error("generate description grpc request error", slog.String("error", err.Error()), slog.String("status", st.Message()))
		return "", InternalDescriptionError
	}
	return description.Description, nil
}

func FullGeneration(ctx context.Context, gRPCClient pb.RamGeneratorClient, userPrompt string, userId int, hybrid bool, descriptions []string, retriesOnNoRam int) (entities.Ram, error) {
	var (
		prompt string
		err    error
	)
	if hybrid {
		prompt, err = GenerateStartPrompt(ctx, gRPCClient, userPrompt)
	} else {
		prompt, err = GenerateHybridPrompt(ctx, gRPCClient, userPrompt, descriptions)
	}
	if err != nil {
		return entities.Ram{}, err
	}

	imageBase64, err := GenerateRamImage(ctx, gRPCClient, prompt)
	if err != nil {
		return entities.Ram{}, err
	}
	imageUrl, err := UploadImage(imageBase64)
	if err != nil {
		return entities.Ram{}, err
	}
	imageDescription, err := GenerateDescription(ctx, gRPCClient, imageUrl)
	if err != nil {
		if errors.Is(err, NoRamError) && retriesOnNoRam > 0 {
			slog.Info("retry after NoRamError", slog.String("place", "FullGeneration"))
			return FullGeneration(ctx, gRPCClient, userPrompt, userId, hybrid, descriptions, retriesOnNoRam-1)
		}
		return entities.Ram{}, err
	}
	return entities.Ram{UserId: userId, Description: imageDescription, ImageUrl: imageUrl}, nil
}
