package views

import "html/template"

func NewView(layout string, files ...string) *View {

	//var files string
	files = append(files, "views/layout/footer.gohtml")
	files = append(files, "views/layout/bootstrap.gohtml")
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}
