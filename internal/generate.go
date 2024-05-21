package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paganotoni/doco/internal/config"
)

// Generates the static html files for the site
// and writes them to the destination folder.
func Generate(srcFolder, dstFolder string, site *site) error {
	// Cleanup the folder
	err := os.RemoveAll(dstFolder)
	if err != nil {
		return err
	}

	// Create the folder
	err = os.MkdirAll(dstFolder, os.ModePerm)
	if err != nil {
		return err
	}

	conf, err := config.Read(srcFolder)
	if err != nil {
		return err
	}

	// Copy assets folder to the destination folder.
	err = copyDir(filepath.Join(srcFolder, "assets"), filepath.Join(dstFolder, "assets"))
	if err != nil {
		return fmt.Errorf("error copying assets: %w", err)
	}

	var pages []generatedPage
	// Generate pages for each of the sections and documents inside them
	// and write them to the destination folder.
	for _, v := range site.sections {
		err := os.MkdirAll(filepath.Join(dstFolder, v.path), os.ModePerm)
		if err != nil {
			return err
		}

		for _, doc := range v.documents {
			// normalize the filename
			name := strings.Replace(doc.filename, filepath.Ext(doc.filename), ".html", 1)
			name = underscore(name)

			data := generatedPage{
				filePath: filepath.Join(dstFolder, v.path, name),

				Title:       doc.title,
				Description: doc.description,
				Keywords:    doc.keywords,

				SectionName: v.name,
				Link:        filepath.Join(v.path, name),

				Content: doc.html,
				Tokens:  doc.Tokens(),

				Navigation: buildNavigation(site, doc),
			}

			if data.Keywords == "" {
				data.Keywords = conf.Keywords
			}

			if data.Description == "" {
				data.Description = conf.Description
			}

			pages = append(pages, data)
		}
	}

	// Generate all of the files after parsing the navigation and
	// having the list to be able to generate the prev and next links.
	for index, v := range pages {
		if index < len(pages)-1 {
			v.Next.Link = pages[index+1].Link
			v.Next.Title = pages[index+1].Title
		}

		if index > 0 {
			v.Prev.Link = pages[index-1].Link
			v.Prev.Title = pages[index-1].Title
		}

		// write the file
		file, err := os.Create(v.filePath)
		if err != nil {
			return err
		}

		err = v.html(conf, file)
		if err != nil {
			return err
		}
	}

	// Generating the site index file to be used by the search.
	f, err := os.Create(filepath.Join(dstFolder, "index.json"))
	if err != nil {
		return fmt.Errorf("error generating search index: %w", err)
	}

	encoder := json.NewEncoder(f)
	err = encoder.Encode(pages)
	if err != nil {
		return fmt.Errorf("error generating search index: %w", err)
	}

	// Adding search.js to the destination folder.
	err = os.WriteFile(filepath.Join(dstFolder, "doco.js"), docoJS, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing search js: %w", err)
	}

	return nil
}
