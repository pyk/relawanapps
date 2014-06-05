package relawanapps

import (
    // "fmt"
    "net/http"
)

func init() {
    http.HandleFunc("/", redirectToRepos)
}

func redirectToRepos(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w,r, "https://github.com/pyk/relawanapps", http.StatusFound)
    return
}