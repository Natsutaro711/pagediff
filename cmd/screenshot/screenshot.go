package screenshot

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func init() {
	err := playwright.Install()
	if err != nil {
		fmt.Println(err)
		return
	}
}

type CsvRow struct {
	Url    string
	Result string
	Path   string
}

type ListCsv struct {
	FileName string
	Rows     []CsvRow
}

func (lc *ListCsv) addRow(record []string) {
	lc.Rows = append(lc.Rows, CsvRow{
		Url:    record[0],
		Result: "",
		Path:   "",
	})
}

func (lc *ListCsv) setRows() error {
	file, err := os.Open(lc.FileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)

	}
	defer file.Close()

	reader := csv.NewReader(file)
	i := 0
	for {
		row, err := reader.Read()
		if i == 0 {
			i++
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		lc.addRow(row)
	}
	return nil
}

func (lc ListCsv) writeResult() error {
	records := [][]string{
		{"URL", "実行結果", "画像パス"},
	}
	for _, row := range lc.Rows {
		records = append(records, []string{
			row.Url,
			row.Result,
			row.Path,
		})
	}
	f, err := os.Create(lc.FileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	w := csv.NewWriter(f)

	w.WriteAll(records)

	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

const ssDir = "screenshots"

func mkCurrentDir() (string, error) {
	n := time.Now().Format("20060102_150405")
	currentDir := path.Join(ssDir, n)

	_, err := os.Stat(currentDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(currentDir, 0755)
		if err != nil {
			return "", fmt.Errorf("could not create ss dir: %v", err)
		}
	}
	return currentDir, nil
}

func defineImageName(u string) (string, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("could not parse url: %v", err)
	}

	h := pu.Host
	h = strings.Replace(h, ".", "-", -1) + "_"

	p := pu.Path
	if p == "" || p == "/" {
		return h + ".png", nil
	}

	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	d, f := path.Split(p)

	ext := path.Ext(f)
	if ext != "" {
		f = strings.TrimSuffix(f, ext) + ".png"
	} else {
		f += ".png"
	}

	d = strings.Replace(d, "/", "", 1)
	d = strings.Replace(d, "/", "-", -1)

	return h + d + f, nil

}

func takePageScreenshot(u string, dir string, br string) (string, error) {
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("could not launch playwright: %v", err)
	}

	bt := pw.Chromium
	if br == "Firefox" {
		bt = pw.Firefox
	} else if br == "WebKit" {
		bt = pw.WebKit
	}

	browser, err := bt.Launch()
	if err != nil {
		return "", fmt.Errorf("could not launch Browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		return "", fmt.Errorf("could not create page: %v", err)
	}
	if _, err = page.Goto(u, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return "", fmt.Errorf("could not goto: %v", err)
	}

	imgName, err := defineImageName(u)
	if err != nil {
		return "", fmt.Errorf("could not define image name: %v", err)
	}

	imgPath := path.Join(dir, imgName)

	if _, err = page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(imgPath),
		FullPage: playwright.Bool(true),
	}); err != nil {
		return "", fmt.Errorf("could not create screenshot: %v", err)
	}
	if err = browser.Close(); err != nil {
		return "", fmt.Errorf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		return "", fmt.Errorf("could not stop Playwright: %v", err)
	}
	return imgPath, nil
}

func ScreenShot(urllist string, browser string) error {
	lc := ListCsv{urllist, []CsvRow{}}
	err := lc.setRows()
	if err != nil {
		return err
	}

	currentDir, err := mkCurrentDir()
	if err != nil {
		return err
	}

	for i := 0; i < len(lc.Rows); i++ {
		imgName, err := takePageScreenshot(lc.Rows[i].Url, currentDir, browser)
		if err != nil {
			fmt.Printf("failed to take page screenshot:%v", err)
			lc.Rows[i].Result = err.Error()
			lc.Rows[i].Path = ""
		}
		lc.Rows[i].Result = "OK"
		lc.Rows[i].Path = imgName
	}

	err = lc.writeResult()
	if err != nil {
		return err
	}

	return nil
}
