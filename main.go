package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"strings"

	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/didip/tollbooth"
	_ "github.com/joho/godotenv/autoload"
	openai "github.com/sashabaranov/go-openai"

	"go.uber.org/zap"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	_ "github.com/joho/godotenv/autoload"
)

var (
	//go:embed templates/index.html
	INDEX_TEMPLATE_TEXT string
	//go:embed templates/paste.html
	PASTE_TEMPLATE_TEXT string
	//go:embed templates/diff.html
	DIFF_TEMPLATE_TEXT string
	//go:embed templates/diff-share.html
	DIFF_SHARED_TEMPLATE_TEXT string
	// embed readme
	//go:embed README.md
	README_TEXT     string
	PBIN_TABLE_NAME = os.Getenv("PBIN_TABLE_NAME")
	PBIN_URL        = os.Getenv("PBIN_URL")
	dataStore       DataStore
)

type PasteTemplateContent struct {
	Text, Language, ID, Title string
}

type DiffTemplateContent struct {
	OldText, NewText string
}

func init() {
	var err error
	dataStore, err = NewDataStore()
	if err != nil {
		log.Fatalf("Failed to initialize data store: %v", err)
	}
}

func generateTitle(text, openapikey string) (string, error) {
	c := openai.NewClient(openapikey)
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `You are a helpful assistant that generates concise, descriptive titles for code snippets or text.
Generate a short, descriptive title (max 10 words) that captures the main purpose or content of the text.
Only output the title, nothing else.`,
			},
			{
				Role:    "user",
				Content: text,
			},
		},
	}
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion generated")
	}
	return resp.Choices[0].Message.Content, nil
}

func handlePaste(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		if err := request.ParseForm(); err != nil {
			fmt.Fprintf(writer, "ParseForm() err: %v", err)
			return
		}
		text := request.FormValue("text")
		lang := request.FormValue("lang")

		// Generate title using OpenAI
		openapikey := os.Getenv("OPENAPIKEY")
		if openapikey == "" {
			log.Printf("OPENAPIKEY not set")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		title, err := generateTitle(text, openapikey)
		if err != nil {
			log.Printf("Failed to generate title: %v", err)
			// Continue without title if generation fails
			title = ""
		}

		id, err := dataStore.AddPaste(text, lang, title)

		if err != nil {
			log.Printf("Failed to add paste: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		q := request.URL.Query()
		q.Del("text")
		q.Del("lang")
		q.Set("id", id)
		request.URL.RawQuery = q.Encode()
		http.Redirect(writer, request, request.URL.String(), http.StatusMovedPermanently)
	case "GET":
		id := request.URL.Query().Get("id")
		if id == "" {
			http.Redirect(writer, request, PBIN_URL, http.StatusMovedPermanently)
			return
		}
		paste, err := dataStore.GetPaste(id)
		if err != nil {
			log.Printf("Failed to get paste: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ptc := PasteTemplateContent{
			Text:     paste.Text,
			Language: paste.Language,
			ID:       id,
			Title:    paste.Title,
		}
		t := template.Must(template.New("paste").Parse(PASTE_TEMPLATE_TEXT))
		err = t.ExecuteTemplate(writer, "paste", ptc)
		if err != nil {
			log.Printf("Failed to execute template: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		http.Redirect(writer, request, PBIN_URL, http.StatusMovedPermanently)
	}
}

type fm struct {
	// title omitempty
	Title *string `yaml:"title,omitempty"`
}

// parseYamlFrontMatter
// check for
// ---
// title: "title"
// ---
// in markdown
//
// returns map[string]string, []byte, error
// map[string]string is the yaml front matter
// []byte is the markdown without the yaml front matter
// error is any error that occured
func parseYamlFrontMatter(md []byte) (*fm, []byte, error) {
	buf := bytes.NewBuffer(md)
	frontMatter := &fm{}
	var err error
	line, err := buf.ReadString('\n')
	if err != nil {
		switch err {
		case io.EOF:
			return nil, md, nil
		default:
			return nil, nil, errors.Wrap(
				err,
				fmt.Sprintf("error reading line: %s", line),
			)
		}
	}
	if !strings.HasPrefix(line, "---") {
		return nil, md, nil
	}

	lines := []string{}
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return nil, nil,
				errors.Wrap(
					err,
					fmt.Sprintf("error reading line: %s", line),
				)
		}
		if strings.HasPrefix(line, "---") {
			break
		}
		lines = append(lines, line)
	}
	y := strings.Join(lines, "")
	err = yaml.Unmarshal([]byte(y), &frontMatter)
	if err != nil {
		return nil, nil, errors.Wrap(
			err,
			fmt.Sprintf("error unmarshalling yaml front matter: %s", y),
		)
	}

	mdWithoutFrontMatter := buf.Bytes()
	return frontMatter, mdWithoutFrontMatter, nil
}

func mdToHTML(md []byte) ([]byte, error) {
	// check for
	// ---
	// title: "title"
	// ---
	// in markdown
	fm, md, err := parseYamlFrontMatter(md)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	maybeUnsafeHTML := markdown.Render(doc, renderer)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
	// add water.css
	// <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">
	// to html
	d, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d.Find("head").AppendHtml(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">`)
	h, err := d.Html()
	if err != nil {
		return nil, err
	}
	// add title if it exists
	if fm != nil && fm.Title != nil {
		d.Find("head").AppendHtml(fmt.Sprintf(`<title>%s</title>`, *fm.Title))
		h, err = d.Html()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return []byte(h), nil

}

func getPaste(id string) (*Paste, error) {
	paste, err := dataStore.GetPaste(id)
	if err != nil {
		return nil, err
	}
	return paste, nil
}

func handleHtml(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		id := request.URL.Query().Get("id")
		if id == "" {
			// redirect to index
			http.Redirect(writer, request, PBIN_URL, http.StatusMovedPermanently)
			return
		}
		paste, err := getPaste(id)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		text := paste.Text
		textBuffer := []byte(text)
		html, err := mdToHTML(textBuffer)
		if err != nil {
			log.Printf("error converting markdown to html, stacktrace: %+v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(html)
		if err != nil {
			log.Println(err)
		}
	default:
		http.Redirect(writer, request, PBIN_URL, http.StatusMovedPermanently)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	_, err := w.Write([]byte(INDEX_TEMPLATE_TEXT))
	if err != nil {
		log.Println(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Println(err)
	}
}

func handleDiff(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		if err := request.ParseForm(); err != nil {
			fmt.Fprintf(writer, "ParseForm() err: %v", err)
			return
		}
		original := request.FormValue("original")
		modified := request.FormValue("modified")
		id, err := dataStore.AddDiff(original, modified)

		if err != nil {
			log.Printf("Failed to add diff: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		q := request.URL.Query()
		q.Del("original")
		q.Del("modified")
		q.Set("id", id)
		request.URL.RawQuery = q.Encode()
		http.Redirect(writer, request, request.URL.String(), http.StatusMovedPermanently)
	case "GET":
		id := request.URL.Query().Get("id")
		if id == "" {
			_, err := writer.Write([]byte(DIFF_TEMPLATE_TEXT))
			if err != nil {
				log.Println(err)
			}
			return
		}
		diff, err := dataStore.GetDiff(id)

		if err != nil {
			log.Printf("Failed to get diff: %v", err)
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		dtc := DiffTemplateContent{
			OldText: diff.OldText,
			NewText: diff.NewText,
		}

		t := template.Must(template.New("diff-share").Parse(DIFF_SHARED_TEMPLATE_TEXT))
		err = t.ExecuteTemplate(writer, "diff-share", dtc)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		http.Redirect(writer, request, PBIN_URL, http.StatusMovedPermanently)
	}
}

func getCompletion(text, openapikey string) ([]string, error) {
	// get completion from  openai

	c := openai.NewClient(openapikey)
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `you are masquerading as github copilot and only provide completions to the text you are given
you only output the completions and do not say anything else. the next message is the text your are given:`,
			},
			{
				Role:    "user",
				Content: text,
			},
		},
	}
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return []string{}, err
	}
	txts := []string{}
	for _, choice := range resp.Choices {
		txts = append(txts, choice.Message.Content)
	}
	return txts, nil

}

type completionResponse struct {
	Completions []string `json:"completions"`
}

func (c completionResponse) ToJsonBytes() ([]byte, error) {
	return json.Marshal(c)
}

func handleCompletion(sugar *zap.SugaredLogger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		openapikey := os.Getenv("OPENAPIKEY")
		// openapikey := "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"
		if openapikey == "" {
			sugar.Error("OPENAPIKEY not set")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch request.Method {
		case "POST":
			if err := request.ParseForm(); err != nil {
				fmt.Fprintf(writer, "ParseForm() err: %v", err)
				return
			}
			text := request.FormValue("text")
			completion, err := getCompletion(text, openapikey)
			sugar.Infow("completion_request", "text", text, "completion", completion, "completion_request", 1)
			if err != nil {
				log.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			resp, err := completionResponse{completion}.ToJsonBytes()
			if err != nil {
				log.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = writer.Write(resp)
			if err != nil {
				log.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			http.Redirect(writer, request, PBIN_URL, http.StatusNotFound)
		}

	}
}

func handleWithDefaultRateLimiter(p string, h http.HandlerFunc) {
	http.Handle(p, tollbooth.LimitFuncHandler(tollbooth.NewLimiter(2, nil), h))
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	handleWithDefaultRateLimiter("/complete", handleCompletion(sugar))
	handleWithDefaultRateLimiter("/diff", handleDiff)
	handleWithDefaultRateLimiter("/health", handleHealth)
	handleWithDefaultRateLimiter("/paste", handlePaste)
	// html
	handleWithDefaultRateLimiter("/html", handleHtml)
	handleWithDefaultRateLimiter("/", handleIndex)
	// get port from env PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	sugar.Infow("starting_server", "port", port)
	sugar.Fatal(http.ListenAndServe(":"+port, nil))
}

