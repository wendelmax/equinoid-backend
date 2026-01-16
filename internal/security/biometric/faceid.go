package biometric

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	mathrand "math/rand"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/security/crypto"
	"github.com/google/uuid"
)

// FaceIDService gerencia autenticação biométrica facial
type FaceIDService struct {
	tolerance         float64
	templateSize      int
	minQuality        float64
	antiSpoofing      bool
	livenessCheck     bool
	encryptionService *crypto.EncryptionService
}

// NewFaceIDService cria um novo serviço Face ID
func NewFaceIDService(tolerance float64, encryptionService *crypto.EncryptionService) *FaceIDService {
	return &FaceIDService{
		tolerance:         tolerance,
		templateSize:      2048, // Tamanho do template biométrico
		minQuality:        0.7,
		antiSpoofing:      true,
		livenessCheck:     true,
		encryptionService: encryptionService,
	}
}

// FaceData representa dados faciais extraídos
type FaceData struct {
	Template    []float64 `json:"template"`
	Quality     float64   `json:"quality"`
	Landmarks   []Point   `json:"landmarks"`
	BoundingBox Rectangle `json:"bounding_box"`
	IsLive      bool      `json:"is_live"`
	Confidence  float64   `json:"confidence"`
}

// Point representa um ponto 2D
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Rectangle representa um retângulo
type Rectangle struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// EnrollmentData representa dados de cadastro biométrico
type EnrollmentData struct {
	UserID        uuid.UUID `json:"user_id"`
	BiometricType string    `json:"biometric_type"`
	ImageData     []byte    `json:"image_data"`
	Quality       float64   `json:"quality"`
}

// VerificationRequest representa uma solicitação de verificação
type VerificationRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	ImageData []byte    `json:"image_data"`
	Challenge string    `json:"challenge,omitempty"`
}

// VerificationResponse representa o resultado da verificação
type VerificationResponse struct {
	Success   bool    `json:"success"`
	Score     float64 `json:"score"`
	Quality   float64 `json:"quality"`
	IsLive    bool    `json:"is_live"`
	Message   string  `json:"message"`
	Challenge string  `json:"challenge,omitempty"`
	SessionID string  `json:"session_id"`
}

// EnrollFace cadastra um template facial para o usuário
func (f *FaceIDService) EnrollFace(data *EnrollmentData) (*models.BiometricData, error) {
	// Extrair dados faciais da imagem
	faceData, err := f.extractFaceData(data.ImageData)
	if err != nil {
		return nil, fmt.Errorf("failed to extract face data: %w", err)
	}

	// Verificar qualidade mínima
	if faceData.Quality < f.minQuality {
		return nil, fmt.Errorf("face quality too low: %.2f (minimum: %.2f)", faceData.Quality, f.minQuality)
	}

	// Verificar liveness (se habilitado)
	if f.livenessCheck && !faceData.IsLive {
		return nil, fmt.Errorf("liveness check failed - possible spoof attempt")
	}

	// Serializar template biométrico
	templateData, err := json.Marshal(faceData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize biometric template: %w", err)
	}

	// Criptografar template biométrico
	encryptedTemplate, err := f.encryptionService.EncryptBytes(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt biometric template: %w", err)
	}

	// Criar registro biométrico
	biometricData := &models.BiometricData{
		UserID:            data.UserID,
		BiometricType:     "face_id",
		BiometricTemplate: encryptedTemplate,
		Quality:           faceData.Quality,
		IsActive:          true,
		EnrollmentDate:    time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return biometricData, nil
}

// VerifyFace verifica uma face contra os templates cadastrados
func (f *FaceIDService) VerifyFace(req *VerificationRequest, storedTemplate []byte) (*VerificationResponse, error) {
	// Descriptografar template armazenado
	decryptedTemplate, err := f.encryptionService.DecryptBytes(storedTemplate)
	if err != nil {
		return &VerificationResponse{
			Success: false,
			Message: "Failed to decrypt stored template",
		}, nil
	}

	// Extrair dados faciais da imagem atual
	currentFace, err := f.extractFaceData(req.ImageData)
	if err != nil {
		return &VerificationResponse{
			Success: false,
			Message: "Failed to extract face data",
		}, nil
	}

	// ... (rest of the checks)

	// Deserializar template descriptografado
	var storedFace FaceData
	if err := json.Unmarshal(decryptedTemplate, &storedFace); err != nil {
		return &VerificationResponse{
			Success: false,
			Message: "Failed to deserialize stored template",
		}, nil
	}

	// Calcular score de similaridade
	score := f.calculateSimilarity(currentFace.Template, storedFace.Template)

	// Verificar se score está acima do threshold
	success := score >= f.tolerance

	response := &VerificationResponse{
		Success:   success,
		Score:     score,
		Quality:   currentFace.Quality,
		IsLive:    currentFace.IsLive,
		SessionID: generateSessionID(),
	}

	if success {
		response.Message = "Face verification successful"
	} else {
		response.Message = fmt.Sprintf("Face verification failed - score: %.3f (required: %.3f)", score, f.tolerance)
	}

	return response, nil
}

// extractFaceData extrai dados faciais de uma imagem (simulado)
func (f *FaceIDService) extractFaceData(imageData []byte) (*FaceData, error) {
	// Simular processamento de imagem (em produção, usar OpenCV, dlib, etc.)
	img, err := f.decodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Simular detecção facial e extração de características
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	// Simular landmarks faciais (pontos característicos)
	landmarks := []Point{
		{X: width * 0.3, Y: height * 0.4}, // Olho esquerdo
		{X: width * 0.7, Y: height * 0.4}, // Olho direito
		{X: width * 0.5, Y: height * 0.6}, // Nariz
		{X: width * 0.3, Y: height * 0.8}, // Canto esquerdo da boca
		{X: width * 0.7, Y: height * 0.8}, // Canto direito da boca
	}

	// Simular bounding box da face
	boundingBox := Rectangle{
		X:      width * 0.2,
		Y:      height * 0.2,
		Width:  width * 0.6,
		Height: height * 0.6,
	}

	// Gerar template biométrico simulado
	template := f.generateFaceTemplate(imageData, landmarks)

	// Simular cálculo de qualidade baseado em resolução e nitidez
	quality := f.calculateQuality(width, height, imageData)

	// Simular liveness detection
	isLive := f.detectLiveness(imageData, landmarks)

	return &FaceData{
		Template:    template,
		Quality:     quality,
		Landmarks:   landmarks,
		BoundingBox: boundingBox,
		IsLive:      isLive,
		Confidence:  quality * 0.95, // Confidence baseada na qualidade
	}, nil
}

// generateFaceTemplate gera um template biométrico (simulado)
func (f *FaceIDService) generateFaceTemplate(imageData []byte, landmarks []Point) []float64 {
	// Simular extração de características faciais
	template := make([]float64, f.templateSize)

	// Usar hash da imagem como seed para gerar template consistente
	hash := sha256.Sum256(imageData)
	seed := int64(0)
	for i := 0; i < 8; i++ {
		seed = seed<<8 + int64(hash[i])
	}

	// Gerar template baseado na seed e landmarks
	for i := 0; i < f.templateSize; i++ {
		// Simular características baseadas em landmarks
		landmarkInfluence := 0.0
		for j, landmark := range landmarks {
			landmarkInfluence += math.Sin(float64(i+j)*0.1) * landmark.X * landmark.Y
		}

		// Combinar com hash da imagem
		hashInfluence := float64(hash[i%32]) / 255.0

		template[i] = (landmarkInfluence + hashInfluence) / 2.0
	}

	return template
}

// calculateQuality calcula a qualidade da imagem facial
func (f *FaceIDService) calculateQuality(width, height float64, imageData []byte) float64 {
	// Simular cálculo de qualidade baseado em:
	// - Resolução da imagem
	// - Tamanho do arquivo (indicativo de compressão)
	// - Distribuição de pixels (indicativo de nitidez)

	resolutionScore := math.Min(1.0, (width*height)/320000.0) // 320x200 como baseline

	fileSizeScore := math.Min(1.0, float64(len(imageData))/50000.0) // 50KB como baseline

	// Simular análise de nitidez através da variação nos dados
	variance := 0.0
	if len(imageData) > 100 {
		for i := 1; i < 100; i++ {
			diff := float64(imageData[i]) - float64(imageData[i-1])
			variance += diff * diff
		}
		variance /= 99.0
	}
	sharpnessScore := math.Min(1.0, variance/1000.0)

	// Combinar scores
	quality := (resolutionScore*0.4 + fileSizeScore*0.3 + sharpnessScore*0.3)

	// Adicionar ruído aleatório para simular variação real
	noise := (float64(mathrand.Intn(200)-100) / 1000.0)
	quality = math.Max(0.0, math.Min(1.0, quality+noise))

	return quality
}

// detectLiveness detecta se a face está viva (anti-spoofing)
func (f *FaceIDService) detectLiveness(imageData []byte, landmarks []Point) bool {
	if !f.livenessCheck {
		return true
	}

	// Simular detecção de liveness através de:
	// - Análise de textura
	// - Detecção de profundidade
	// - Análise de movimento (em sequências)
	// - Detecção de reflexão ocular

	// Simular análise de textura
	textureScore := 0.0
	if len(imageData) > 1000 {
		for i := 0; i < 1000; i += 10 {
			textureScore += math.Abs(float64(imageData[i]) - float64(imageData[i+5]))
		}
		textureScore /= 100.0
	}

	// Simular detecção de profundidade baseada em landmarks
	depthScore := 0.0
	if len(landmarks) >= 5 {
		// Calcular distâncias entre landmarks para detectar "planitude"
		eyeDistance := math.Sqrt(math.Pow(landmarks[1].X-landmarks[0].X, 2) +
			math.Pow(landmarks[1].Y-landmarks[0].Y, 2))
		noseToMouthDistance := math.Sqrt(math.Pow(landmarks[2].X-landmarks[3].X, 2) +
			math.Pow(landmarks[2].Y-landmarks[3].Y, 2))

		// Face real deve ter proporções específicas
		if eyeDistance > 0 && noseToMouthDistance > 0 {
			ratio := eyeDistance / noseToMouthDistance
			if ratio > 0.8 && ratio < 2.5 {
				depthScore = 1.0
			}
		}
	}

	// Combinar scores
	livenessScore := (textureScore*0.6 + depthScore*0.4) / 255.0

	// Face é considerada viva se score > 0.5
	return livenessScore > 0.5
}

// calculateSimilarity calcula a similaridade entre dois templates
func (f *FaceIDService) calculateSimilarity(template1, template2 []float64) float64 {
	if len(template1) != len(template2) || len(template1) == 0 {
		return 0.0
	}

	// Calcular distância euclidiana normalizada
	sumSquares := 0.0
	for i := 0; i < len(template1); i++ {
		diff := template1[i] - template2[i]
		sumSquares += diff * diff
	}

	distance := math.Sqrt(sumSquares)
	maxDistance := math.Sqrt(float64(len(template1)) * 4.0) // Assumindo valores entre -2 e 2

	// Converter distância em score de similaridade (0-1)
	similarity := math.Max(0.0, 1.0-distance/maxDistance)

	return similarity
}

// decodeImage decodifica dados de imagem
func (f *FaceIDService) decodeImage(imageData []byte) (image.Image, error) {
	reader := bytes.NewReader(imageData)
	img, _, err := image.Decode(reader)
	if err != nil {
		// Tentar JPEG especificamente
		reader.Seek(0, 0)
		img, err = jpeg.Decode(reader)
	}
	return img, err
}

// generateSessionID gera um ID de sessão único
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// GetSupportedFormats retorna os formatos de imagem suportados
func (f *FaceIDService) GetSupportedFormats() []string {
	return []string{"JPEG", "PNG", "BMP", "GIF"}
}

// GetRecommendedSettings retorna configurações recomendadas para captura
func (f *FaceIDService) GetRecommendedSettings() map[string]interface{} {
	return map[string]interface{}{
		"min_resolution": "640x480",
		"max_file_size":  "5MB",
		"format":         "JPEG",
		"quality":        85,
		"lighting":       "well_lit",
		"angle":          "frontal",
		"distance":       "60-100cm",
	}
}
