package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/fatih/color"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
)

type Result struct {
	Title string
	Link  string
}

func fetchResults(query string) []Result {
	q := strings.ReplaceAll(query, " ", "+")
	searchURL := "https://html.duckduckgo.com/html/?q=" + q

	res, err := http.Get(searchURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var results []Result
	doc.Find(".result__title").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= 10 {
			return false
		}
		title := strings.TrimSpace(s.Text())
		href, _ := s.Find("a").Attr("href")
		u, _ := url.Parse(href)
		link := u.Query().Get("uddg")
		results = append(results, Result{Title: title, Link: link})
		return true
	})
	return results
}

func extractInfo(link string, tag string) []string {
	res, err := http.Get(link)
	if err != nil {
		log.Println("Hata:", err)
		return nil
	}
	defer res.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(res.Body)
	var content []string

	doc.Find(tag).EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			content = append(content, text)
		}
		return true
	})
	return content
}

func extractMeta(link string) (string, string) {
	res, err := http.Get(link)
	if err != nil {
		return "", ""
	}
	defer res.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(res.Body)
	title := doc.Find("title").First().Text()
	desc, _ := doc.Find("meta[name='description']").Attr("content")
	return title, desc
}

func saveToFile(filename string, content []string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Println("Dosya oluşturulamadı:", err)
		return
	}
	defer f.Close()
	for _, line := range content {
		f.WriteString(line + "\n")
	}
	color.Green("💾 İçerik '%s' dosyasına kaydedildi.\n", filename)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		color.Cyan("Aramak istediğiniz kelimeleri girin (virgülle ayırarak): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		results := fetchResults(input)

		color.Yellow("\n🔍 İlk 10 sonuç:")
		for i, r := range results {
			color.White("%d. %s", i+1, r.Title)
			color.Blue("   Link: %s\n", r.Link)
		}

		fmt.Print("\nİşlem yapmak istiyor musunuz? (y/n/h): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "n" {
			break
		} else if choice == "h" {
			continue
		} else if choice == "y" {
			fmt.Print("İşlem yapmak istediğiniz sonucu seçin (1-10): ")
			numInput, _ := reader.ReadString('\n')
			numInput = strings.TrimSpace(numInput)
			index := int(numInput[0] - '1')

			if index < 0 || index >= len(results) {
				color.Red("❌ Geçersiz seçim!")
				continue
			}

			link := results[index].Link

			fmt.Println("\nNe çekmek istiyorsunuz?")
			fmt.Println("1. İlk 5 <p> etiketi")
			fmt.Println("2. <h1> etiketi")
			fmt.Println("3. Tüm <h*> etiketleri")
			fmt.Println("4. Tüm <div> etiketleri")
			fmt.Println("5. Belirli bir class adı")
			fmt.Println("6. <title> ve <meta description>")
			fmt.Print("Seçiminiz (1-6): ")

			opt, _ := reader.ReadString('\n')
			opt = strings.TrimSpace(opt)

			var content []string

			switch opt {
			case "1":
				content = extractInfo(link, "p")
				if len(content) > 5 {
					content = content[:5]
				}
			case "2":
				content = extractInfo(link, "h1")
			case "3":
				for i := 1; i <= 6; i++ {
					content = append(content, extractInfo(link, fmt.Sprintf("h%d", i))...)
				}
			case "4":
				content = extractInfo(link, "div")
			case "5":
				fmt.Print("Class adını girin: ")
				className, _ := reader.ReadString('\n')
				className = strings.TrimSpace(className)
				content = extractInfo(link, "."+className)
			case "6":
				title, desc := extractMeta(link)
				content = append(content, "Title: "+title)
				content = append(content, "Meta Açıklama: "+desc)
			default:
				color.Red("❌ Geçersiz seçim!")
				continue
			}

			color.Magenta("\n📄 İçerik:")
			for _, c := range content {
				color.White("- " + c)
			}

			fmt.Print("\nİçeriği dosyaya kaydetmek ister misin? (y/n): ")
			saveChoice, _ := reader.ReadString('\n')
			saveChoice = strings.TrimSpace(saveChoice)

			if saveChoice == "y" {
				fmt.Print("Dosya adı girin (örnek: sonuc.txt): ")
				filename, _ := reader.ReadString('\n')
				filename = strings.TrimSpace(filename)
				saveToFile(filename, content)
			}
		}
	}
}
