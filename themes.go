/* Copyright Azareal 2016 - 2017 */
package main

import "fmt"
import "log"
import "io"
import "os"
import "strings"
import "mime"
import "io/ioutil"
import "path/filepath"
import "encoding/json"
import "net/http"

var defaultTheme string
var themes map[string]Theme = make(map[string]Theme)
//var overriden_templates map[string]interface{} = make(map[string]interface{})
var overriden_templates map[string]bool = make(map[string]bool)

type Theme struct
{
	Name string
	FriendlyName string
	Version string
	Creator string
	Settings map[string]ThemeSetting
	Templates []TemplateMapping
	
	// This variable should only be set and unset by the system, not the theme meta file
	Active bool
}

type ThemeSetting struct
{
	FriendlyName string
	Options []string
}

type TemplateMapping struct
{
	Name string
	Source string
	//When string
}

func init_themes() {
	themeFiles, err := ioutil.ReadDir("./themes")
	if err != nil {
		log.Fatal(err)
	}
	
	for _, themeFile := range themeFiles {
		if !themeFile.IsDir() {
			continue
		}
		
		themeName := themeFile.Name()
		log.Print("Adding theme '" + themeName + "'")
		themeFile, err := ioutil.ReadFile("./themes/" + themeName + "/theme.json")
		if err != nil {
			log.Fatal(err)
		}
		
		var theme Theme
		json.Unmarshal(themeFile, &theme)
		theme.Active = false // Set this to false, just in case someone explicitly overrode this value in the JSON file
		
		themes[theme.Name] = theme
	}
}

func add_theme_static_files(themeName string) {
	err := filepath.Walk("./themes/" + themeName + "/public", func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		path = strings.Replace(path,"\\","/",-1)
		
		log.Print("Attempting to add static file '" + path + "' for default theme '" + themeName + "'")
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		
		path = strings.TrimPrefix(path,"themes/" + themeName + "/public")
		log.Print("Added the '" + path + "' static file for default theme " + themeName + ".")
		
		static_files["/static" + path] = SFile{data,0,int64(len(data)),mime.TypeByExtension(filepath.Ext("/themes/" + themeName + "/public" + path)),f,f.ModTime().UTC().Format(http.TimeFormat)}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func map_theme_templates(theme Theme) {
	if theme.Templates != nil {
		for _, themeTmpl := range theme.Templates {
			if themeTmpl.Name == "" {
				log.Fatal("Invalid destination template name")
			}
			if themeTmpl.Source == "" {
				log.Fatal("Invalid source template name")
			}
			
			// go generate is one possibility, but it would simply add another step of compilation
			
			dest_tmpl_ptr, ok := tmpl_ptr_map[themeTmpl.Name]
			if !ok {
				log.Fatal("The destination template doesn't exist!")
			}
			source_tmpl_ptr, ok := tmpl_ptr_map[themeTmpl.Source]
			if !ok {
				log.Fatal("The source template doesn't exist!")
			}
			
			switch d_tmpl_ptr := dest_tmpl_ptr.(type) {
				case *func(TopicPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(TopicPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							log.Print("Topic Handle")
							fmt.Println(template_topic_handle)
							log.Print("Before")
							fmt.Println(d_tmpl_ptr)
							fmt.Println(*d_tmpl_ptr)
							log.Print("Source")
							fmt.Println(s_tmpl_ptr)
							fmt.Println(*s_tmpl_ptr)
							*d_tmpl_ptr = *s_tmpl_ptr
							log.Print("After")
							fmt.Println(d_tmpl_ptr)
							fmt.Println(*d_tmpl_ptr)
							log.Print("Source")
							fmt.Println(s_tmpl_ptr)
							fmt.Println(*s_tmpl_ptr)
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(TopicsPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(TopicsPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(ForumPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(ForumPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(ForumsPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(ForumsPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(ProfilePage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(ProfilePage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(Page,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(Page,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				default:
					log.Fatal("Unknown destination template type!")
			}
		}
	}
}

func reset_template_overrides() {
	log.Print("Resetting the template overrides")
	
	for name, _ := range overriden_templates {
		log.Print("Resetting '" + name + "' template override")
		
		origin_pointer, ok := tmpl_ptr_map["o_" + name]
		if !ok {
			//log.Fatal("The origin template doesn't exist!")
			log.Print("The origin template doesn't exist!")
			return
		}
		
		dest_tmpl_ptr, ok := tmpl_ptr_map[name]
		if !ok {
			//log.Fatal("The destination template doesn't exist!")
			log.Print("The destination template doesn't exist!")
			return
		}
		
		switch o_ptr := origin_pointer.(type) {
			case func(TopicPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(TopicPage,io.Writer):
						log.Print("Topic Handle")
						fmt.Println(template_topic_handle)
						log.Print("Before")
						fmt.Println(d_ptr)
						fmt.Println(*d_ptr)
						log.Print("Origin")
						fmt.Println(o_ptr)
						*d_ptr = o_ptr
						log.Print("After")
						fmt.Println(d_ptr)
						fmt.Println(*d_ptr)
						log.Print("Origin")
						fmt.Println(o_ptr)
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(TopicsPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(TopicsPage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(ForumPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ForumPage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(ForumsPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ForumsPage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(ProfilePage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ProfilePage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			default:
				log.Fatal("Unknown destination template type!")
		}
		log.Print("The template override was reset")
	}
	overriden_templates = make(map[string]bool)
	log.Print("All of the template overrides have been reset")
	
	/*for name, origin_pointer := range overriden_templates {
		dest_tmpl_ptr, ok := tmpl_ptr_map[name]
		if !ok {
			log.Fatal("The destination template doesn't exist!")
		}
		
		switch o_ptr := origin_pointer.(type) {
			case *func(TopicPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(TopicPage,io.Writer):
						*d_ptr = *o_ptr
					default:
							log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(TopicsPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(TopicsPage,io.Writer):
						*d_ptr = *o_ptr
					default:
							log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(ForumPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ForumPage,io.Writer):
						*d_ptr = *o_ptr
					default:
							log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(ForumsPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ForumsPage,io.Writer):
						*d_ptr = *o_ptr
					default:
							log.Fatal("The origin and destination templates are incompatible")
				}
			case *func(ProfilePage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ProfilePage,io.Writer):
						*d_ptr = *o_ptr
					default:
							log.Fatal("The origin and destination templates are incompatible")
				}
			default:
				log.Fatal("Unknown destination template type!")
		}
		delete(overriden_templates, name)
	}*/
}