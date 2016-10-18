package views

import "html/template"

func NewView(files ...string) *View {

	//var files string
	files = append(files, "views/layout/footer.gohtml")
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
}
