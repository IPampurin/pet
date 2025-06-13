package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Hello Fyne")

	label := widget.NewLabel("Привет, Fyne!")
	myWindow.SetContent(container.NewVBox(label))

	myWindow.ShowAndRun()
}
