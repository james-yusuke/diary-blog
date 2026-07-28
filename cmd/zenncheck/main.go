// Command zenncheck reports every Zenn article's publication state and makes
// sure all non-draft articles are present in the Worker generation input.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/james-yusuke/diary-blog/internal/content"
)

func main() {
	root := flag.String("root", ".", "project root")
	flag.Parse()
	if err := check(*root, time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(root string, now time.Time) error {
	contentDir := filepath.Join(root, "content")
	store, err := content.LoadAll(filepath.Join(root, "config", "site.yaml"), contentDir)
	if err != nil {
		return err
	}
	audit, err := content.AuditZenn(contentDir, now)
	if err != nil {
		return err
	}

	inSnapshot := make(map[string]bool)
	for _, post := range store.Posts {
		if post.IsZenn() {
			inSnapshot[post.Slug] = true
		}
	}

	counts := make(map[content.ZennArticleState]int)
	for _, article := range audit {
		counts[article.State]++
		switch article.State {
		case content.ZennArticleDraft:
			fmt.Printf("DRAFT     %s (intentionally excluded)\n", article.Slug)
		case content.ZennArticlePublic, content.ZennArticleScheduled:
			if !inSnapshot[article.Slug] {
				return fmt.Errorf("Zenn article %q is %s but missing from the Worker snapshot input", article.Slug, article.State)
			}
			if article.State == content.ZennArticleScheduled {
				fmt.Printf("SCHEDULED %s (%s)\n", article.Slug, article.Published.Format(time.RFC3339))
			} else {
				fmt.Printf("PUBLIC    %s\n", article.Slug)
			}
		}
	}
	fmt.Printf("Zenn coverage: %d public, %d scheduled, %d draft\n", counts[content.ZennArticlePublic], counts[content.ZennArticleScheduled], counts[content.ZennArticleDraft])
	return nil
}
