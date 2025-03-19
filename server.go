package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"os/exec"
	"strconv"
)

//////// SETTINGS ///////
var rotation_time float64 = 0 // in seconds
/////////////////////////

var curr_wallpaper_index int
var color_ref map[string][]int
var wallpapers []string	
var curr_wallpaper string
var last_time time.Time

func check(e error) {if e != nil {panic(e)}}

func get_color_ref() map[string][]int {
	dat, err := os.ReadFile("/Users/river/.config/scripts/color_reference.txt")
	check(err)
	sv := strings.Split(string(dat),"\n")
	cref := make(map[string][]int)
	for _, line := range sv[:len(sv)-1] {
		splitv := strings.Split(line,";")
		cmap := strings.Replace(splitv[1],","," ",2);
		colmap := make([]int, 3)
		for i, color := range strings.Split(cmap," ") {
			colmap[i], _ = strconv.Atoi(color)
		}
		cref[splitv[0]] = colmap
	}
	return cref
}

func page_wallpapers() {
	wallpapers = get_wallpapers()
	color_ref = get_color_ref()
}

func get_delay() float64 {
	dat, err := os.ReadFile("/Users/river/.config/wallpaper_server/fdelay.txt")
	check(err)
	delay, _ := strconv.Atoi(string(dat)[:len(string(dat))-1])
	return float64(delay)
}

func current_wallpaper() string {
	out, err := exec.Command("osascript", "-e", "tell app \"finder\" to get posix path of (get desktop picture as alias)").Output()
	check(err)
	return string(out)
	// RUNTIME: ~270ms, less File I/O than set_wallpaper
}

func set_wallpaper(wallpaper string) {
	_, err := exec.Command("osascript", "-e", "tell app \"finder\" to set desktop picture to POSIX file \""+wallpaper+"\"").Output()
	check(err)
	// RUNTIME ~270ms
}
func get_wallpapers() []string {
	files, err := os.ReadDir("/Users/river/.config/assets/wallpapers")
	check(err)
	wallpapers := make([]string, len(files))
	var builder strings.Builder
	for i, file := range files {
		builder.WriteString("/Users/river/.config/assets/wallpapers/")
		builder.WriteString(file.Name())
		wallpapers[i] = builder.String()
		builder.Reset()
	}
	// sort.Strings(wallpapers)
	// http://thecodelesscode.com/case/43
	// MacOS is HFS+ traversal. It's guaranteed to be sorted (or at least guaranteed to be in the same order every time, which is all we actually want).
	return wallpapers
}

func handler_getwpaper(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, current_wallpaper())
}

func handler_setwpaper(w http.ResponseWriter, r *http.Request) {
	set_wallpaper(curr_wallpaper)
}

func get_color() string {
	var sb strings.Builder
	cr := color_ref[curr_wallpaper]
	for i, color := range cr {
		sb.WriteString(strconv.Itoa(color))
		if i != len(cr)-1 {
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

func fix_sketchybar() {
	var hex_color strings.Builder
	for _, color := range color_ref[curr_wallpaper] {
		hex_color.WriteString(strconv.FormatInt(int64(color), 16))
	}
	_, err := exec.Command("/bin/sh","-c","/usr/local/bin/sketchybar --set spaces background.border_color=\"0xff"+hex_color.String()+"\"").Output()
	check(err)
}

func handler_gcol(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, get_color())
}

func rotate_wallpaper() {
	if delta_compare() {
		change_wallpaper()
	}
    if current_wallpaper() != curr_wallpaper {
		set_wallpaper(curr_wallpaper)
	}
}

func change_wallpaper() {
	curr_wallpaper_index = (curr_wallpaper_index+1)%len(wallpapers)
	curr_wallpaper = wallpapers[curr_wallpaper_index]
	set_wallpaper(curr_wallpaper)
	fix_sketchybar()
	last_time = time.Now()
}

func handle_changed(w http.ResponseWriter, r *http.Request) {
	change_wallpaper()
	fmt.Fprintf(w, "OK\n\r")
}

func handle_rotate(w http.ResponseWriter, r *http.Request) {
	rotate_wallpaper()
}

func handle_page(w http.ResponseWriter, r *http.Request) {
	page_wallpapers()
	fmt.Fprintf(w, "Paged!\n\r")
}

func delta_compare() bool {
	return time.Since(last_time).Seconds() > rotation_time
}

func main() {
	rotation_time = get_delay()
	color_ref = get_color_ref()
	wallpapers = get_wallpapers()
	curr_wallpaper = current_wallpaper()
	for i, wallpaper := range wallpapers {
		if wallpaper == curr_wallpaper {
			curr_wallpaper_index = i
			break
		}
	}
	last_time = time.Now()
	http.HandleFunc("/get_wallpaper", handler_getwpaper)
	http.HandleFunc("/set_bg", handler_setwpaper)
	http.HandleFunc("/get_color", handler_gcol)
	http.HandleFunc("/change_wallpaper", handle_changed)
	http.HandleFunc("/rotate", handle_rotate)
	http.HandleFunc("/page", handle_page)
	err := http.ListenAndServe(":6969", nil)
	fmt.Println(err)
}

