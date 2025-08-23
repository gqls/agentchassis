// FILE: platform/orchestration/actions/image_actions.go
package actions

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/disintegration/imaging"
)

// GenerateImageAction generates images using AI
func GenerateImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Generating image")

	// Get image generation config
	config := params.StepConfig.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	// Build prompt from context
	prompt := buildImagePrompt(params.CollectedData, config)

	// Get image generation parameters
	model := getStringOrDefault(config, "model", "dall-e-3")
	size := getStringOrDefault(config, "size", "1024x1024")
	quality := getStringOrDefault(config, "quality", "standard")
	style := getStringOrDefault(config, "style", "vivid")

	// Call image generation API (placeholder - would integrate with actual API)
	imageData, err := callImageGenerationAPI(ctx, prompt, model, size, quality, style)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	return map[string]interface{}{
		"generated_image": imageData,
		"prompt":          prompt,
		"model":           model,
		"size":            size,
		"format":          "png",
	}, nil
}

// ProcessImageAction processes and optimizes images
func ProcessImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Processing image")

	// Get image data
	var imageData []byte

	if genResult, ok := params.CollectedData["generate_image"].(map[string]interface{}); ok {
		imageData = extractImageData(genResult["generated_image"])
	} else if imgData, ok := params.CollectedData["image"]; ok {
		imageData = extractImageData(imgData)
	}

	if len(imageData) == 0 {
		return nil, fmt.Errorf("no image data found")
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get processing config
	config := params.StepConfig.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	// Apply processing
	processedImg := img
	processingSteps := []string{}

	// Resize if needed
	if width, ok := config["resize_width"].(float64); ok {
		processedImg = imaging.Resize(processedImg, int(width), 0, imaging.Lanczos)
		processingSteps = append(processingSteps, fmt.Sprintf("resize_width_%d", int(width)))
	}

	// Apply filters
	if brightness, ok := config["brightness"].(float64); ok {
		processedImg = imaging.AdjustBrightness(processedImg, brightness)
		processingSteps = append(processingSteps, "brightness_adjust")
	}

	if contrast, ok := config["contrast"].(float64); ok {
		processedImg = imaging.AdjustContrast(processedImg, contrast)
		processingSteps = append(processingSteps, "contrast_adjust")
	}

	if config["sharpen"] == true {
		processedImg = imaging.Sharpen(processedImg, 1.0)
		processingSteps = append(processingSteps, "sharpen")
	}

	// Add watermark if configured
	if watermark, ok := config["watermark"].(string); ok && watermark != "" {
		processedImg = addWatermark(processedImg, watermark)
		processingSteps = append(processingSteps, "watermark")
	}

	// Encode processed image
	var outputBuf bytes.Buffer
	outputFormat := getStringOrDefault(config, "output_format", format)

	switch outputFormat {
	case "jpeg", "jpg":
		quality := 90
		if q, ok := config["jpeg_quality"].(float64); ok {
			quality = int(q)
		}
		err = jpeg.Encode(&outputBuf, processedImg, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&outputBuf, processedImg)
	default:
		err = png.Encode(&outputBuf, processedImg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return map[string]interface{}{
		"processed_image": outputBuf.Bytes(),
		"format":          outputFormat,
		"original_size":   len(imageData),
		"processed_size":  outputBuf.Len(),
		"dimensions": map[string]int{
			"width":  processedImg.Bounds().Dx(),
			"height": processedImg.Bounds().Dy(),
		},
		"processing_steps": processingSteps,
	}, nil
}

// ValidateImageAction validates image data
func ValidateImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Validating image")

	// Get image data
	var imageData []byte

	if procResult, ok := params.CollectedData["process_image"].(map[string]interface{}); ok {
		imageData = extractImageData(procResult["processed_image"])
	} else if imgData, ok := params.CollectedData["image"]; ok {
		imageData = extractImageData(imgData)
	}

	if len(imageData) == 0 {
		return map[string]interface{}{
			"valid":  false,
			"errors": []string{"No image data found"},
		}, nil
	}

	// Decode and validate
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return map[string]interface{}{
			"valid":  false,
			"errors": []string{fmt.Sprintf("Failed to decode image: %v", err)},
		}, nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	errors := []string{}
	warnings := []string{}

	// Check dimensions
	config := params.StepConfig.Config
	if minWidth, ok := config["min_width"].(float64); ok && width < int(minWidth) {
		errors = append(errors, fmt.Sprintf("Image width %d is less than minimum %d", width, int(minWidth)))
	}

	if minHeight, ok := config["min_height"].(float64); ok && height < int(minHeight) {
		errors = append(errors, fmt.Sprintf("Image height %d is less than minimum %d", height, int(minHeight)))
	}

	if maxWidth, ok := config["max_width"].(float64); ok && width > int(maxWidth) {
		warnings = append(warnings, fmt.Sprintf("Image width %d exceeds maximum %d", width, int(maxWidth)))
	}

	if maxHeight, ok := config["max_height"].(float64); ok && height > int(maxHeight) {
		warnings = append(warnings, fmt.Sprintf("Image height %d exceeds maximum %d", height, int(maxHeight)))
	}

	// Check file size
	if maxSize, ok := config["max_size_bytes"].(float64); ok && len(imageData) > int(maxSize) {
		warnings = append(warnings, fmt.Sprintf("Image size %d bytes exceeds maximum %d", len(imageData), int(maxSize)))
	}

	isValid := len(errors) == 0

	// Store final image if valid
	if isValid {
		params.CollectedData["final_image"] = imageData
	}

	return map[string]interface{}{
		"valid":    isValid,
		"errors":   errors,
		"warnings": warnings,
		"format":   format,
		"dimensions": map[string]int{
			"width":  width,
			"height": height,
		},
		"size_bytes":  len(imageData),
		"final_image": imageData,
	}, nil
}

// Helper functions

func buildImagePrompt(collectedData map[string]interface{}, config map[string]interface{}) string {
	// Extract context for image generation
	var promptParts []string

	if prompt, ok := config["prompt"].(string); ok {
		promptParts = append(promptParts, prompt)
	}

	// Add business context if available
	if businessInfo, ok := collectedData["business_info"].(map[string]interface{}); ok {
		if name, ok := businessInfo["business_name"].(string); ok {
			promptParts = append(promptParts, fmt.Sprintf("for %s", name))
		}
	}

	// Add style preferences
	if style, ok := config["style"].(string); ok {
		promptParts = append(promptParts, fmt.Sprintf("in %s style", style))
	}

	if len(promptParts) == 0 {
		return "Generate a professional business image"
	}

	return strings.Join(promptParts, " ")
}

func extractImageData(data interface{}) []byte {
	switch v := data.(type) {
	case []byte:
		return v
	case string:
		// Check if it's base64 encoded
		if strings.HasPrefix(v, "data:image") {
			parts := strings.Split(v, ",")
			if len(parts) == 2 {
				decoded, _ := base64.StdEncoding.DecodeString(parts[1])
				return decoded
			}
		}
		// Otherwise treat as raw bytes
		return []byte(v)
	default:
		return nil
	}
}

func addWatermark(img image.Image, watermarkText string) image.Image {
	// This would add a watermark to the image
	// Implementation would use golang.org/x/image/font and draw
	return img
}

func callImageGenerationAPI(ctx context.Context, prompt, model, size, quality, style string) ([]byte, error) {
	// This would call the actual image generation API (OpenAI, Stability, etc.)
	// For now, return placeholder
	return []byte("placeholder_image_data"), nil
}
