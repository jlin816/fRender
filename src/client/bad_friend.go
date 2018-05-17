package client

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (fr *Friend) badRenderFrames(file string, frames []int) string {
	relativeFolder := fr.getLocalFilename(fmt.Sprintf("%v_frames_%v", file, fr.username))
	penguin := "files/bad/penguin.png"
	if _, err := os.Stat(relativeFolder); os.IsNotExist(err) {
		os.Mkdir(relativeFolder, os.ModePerm)
	}

	for _, frame := range frames {
		outFile := fmt.Sprintf("%v/frame_%05d.png", relativeFolder, frame)

		cpCmd := exec.Command("/bin/cp", "-f", penguin, outFile)
		// fmt.Printf("%v %v %v %v\n", "/bin/cp", "-f", penguin, outFile)
		err := cpCmd.Run()
		if err != nil {
			panic(err)
		}
		fr.logImg(fmt.Sprintf("%v_frames_%v/frame_%05d.png", file, fr.username, frame))

	}

	zipCmd := exec.Command("zip", "-rj", relativeFolder+".zip", relativeFolder)
	err1 := zipCmd.Run()
	if err1 != nil {
		panic(err1)
	}
	time.Sleep(100 * time.Millisecond)
	os.RemoveAll(relativeFolder)
	return fmt.Sprintf("%v_frames_%v", file, fr.username) + ".zip"
}
