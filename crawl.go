/**
 * open directory to m3u generator
 * Simple application to recurse through an open directory, download the list of files and make a playlist from it
 *
 * build with: go build opendirectorytom3u.go
 */
 
 // fixme urls with # in will cause it to get the root url - why?
// todo i think i need to ascertain the length of the file where possible as it can cause the end of the file to be cut short
// todo optional parsing of filename-to-title and remove extension and convert "." to spaces
// todo use go subroutines to recurse through the directories faster?
 
package main
 
import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "path/filepath"
    "regexp"
    "strings"
)
 
func main() {
    println(" ** Open Directory To M3U Generator **\n")
 
    // get open directory url
    println("Enter the URL of the open directory, include http(s):// and omit trailing slash:")
    baseURL := getUserInput()
 
    // ensure the url starts with http
    if baseURL[:4] != "http" {
        log.Fatalln("URL does not start with http")
    }
 
    // remove trailing slash incase the user is a moron
    if baseURL[len(baseURL) - 1:] == "/" {
        baseURL = baseURL[:len(baseURL) - 1]
    }
 
    println("Gathering files and directories, please wait...")
 
    // download the page
    c := new(http.Client)
 
    urls := parseURLRecursively(baseURL, c)
 
    println("\nFinished parsing, let's write it to an M3U!")
    println("Enter the output filename of the m3u")
    fn := getUserInput()
 
    CreateM3U(baseURL, urls, fn)
 
    println("All done!")
}
 
// generate the m3u file from a list of urls to files
func CreateM3U(baseURL string, urls []string, outputFileName string) {
 
    // m3u header
    m3u := "#EXTM3U\n"
 
    // add items to the m3u
    for _, u := range urls {
        // the -1 here denotes the length of the file, we don't know it, so -1 means 'not known'
        encoded, _ := url.PathUnescape(u)
        m3u += fmt.Sprintf("#EXTINF:-1, %s\n", encoded)
        m3u += baseURL + "/" + u + "\n"
    }
 
    // write the file to the specified output directory
    ioutil.WriteFile(outputFileName, []byte(m3u), 0777)
}
 
// parse an url and pull out directories and files
func parseURLRecursively(theURL string, c *http.Client) []string {
 
    var urls []string
 
    s, err := url.PathUnescape(theURL)
    resp, err := c.Get(s)
    if err != nil {
        log.Fatalln(err)
    }
 
    // close the body of the current html response at the end of the function
    defer resp.Body.Close()
 
    // parse the html for files and recurse through directories
    var directories []string
    var files []string
 
    if resp.StatusCode == http.StatusOK {
 
        // extract the html from the response
        htmlBytes, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Fatalln(err)
        }
        html := string(htmlBytes)
 
        // get files
        files = parseFiles(html)
 
        // add files to the list of urls
        for _,f := range files {
            urls = append(urls, f)
        }
 
        // get directories
        directories = parseDirectories(html)
 
        for _, dir := range directories {
 
            s, _ := url.PathUnescape(dir)
 
            // fixme for some reason, urls with # screw up, so we skip them for now
            if strings.Contains(s, "#") {
                continue
            }
 
            println(s)
 
            // get the new page
            // todo is there a // instead of a / here?
            subdirURL := theURL + "/" + dir
 
            newurls := parseURLRecursively(subdirURL, c)
            if len(newurls) > 0 {
 
                // todo is it possible to merge two arrays together rather than doing this?
                for _, u := range newurls {
                    urls = append(urls, dir + "/" + u)
                }
            }
        }
    }
 
    return urls
}
 
// parse a directory recursively
func parseDirectories(html string) []string {
 
    var dirs []string
 
    // we're using regex! so sue me!
    reg, err := regexp.Compile("<a href=\"([^\"]+)/\">[^\"]+</a>")
    if err != nil {
        log.Fatalln(err)
    }
 
    matches := reg.FindAllStringSubmatch(html, -1)
    if matches == nil {
        // no directories found
        return dirs
    }
 
//  println(fmt.Sprintf("Found %d sub-directories", len(matches)))
 
    for _, match := range matches {
        m := match[1]
 
        dirs = append(dirs, m)
    }
 
    return dirs
}
 
// parse files
func parseFiles(html string) []string {
 
    // only files with these extension will be added to the playlist output
    // note the period is necessary as the first character
    whitelistExtensions := []string{
        ".mp3", ".mp4", ".flac", ".wmv", ".m4b", ".mov",
        // ".mkv", // files that can't be streamed yet
    }
 
    var files []string
 
    // we're using regex! so sue me!
    reg, err := regexp.Compile("<a href=\"([^\"]+)\">[^\"]+</a>")
    if err != nil {
        log.Fatalln(err)
    }
 
    matches := reg.FindAllStringSubmatch(html, -1)
    if matches == nil {
        return files
//      log.Fatalln("no files found")
    }
 
//  println(fmt.Sprintf("Found %d matches", len(matches)))
 
    for _, match := range matches {
        m := match[1]
 
        // the first character of the extension returned from .Ext is the "."
        ext := filepath.Ext(m)
 
        // check the extension is in the above extension whitelist
        for _,e := range whitelistExtensions {
 
            // if found, then it's an extension we accept and we add it to the list of files
            if ext == e {
                print(".")
//              println(m)
                files = append(files, m)
                break
            }
        }
    }
 
    return files
}
 
// helper to get input from user
func getUserInput() string {
    var inp string
    fmt.Scanln(&inp)
    return inp
}
