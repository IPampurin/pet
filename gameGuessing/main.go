package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Hello Fyne")

	label := widget.NewLabel("Привет, Fyne!")
	myWindow.SetContent(container.NewVBox(label))

	myWindow.ShowAndRun()
}
