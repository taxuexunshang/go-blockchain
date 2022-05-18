package main

func main() {
	// 这个地方会存在一个坑 在通过命令运行的时候要将所有涉及调用的文件都加上 或者编译之后再运行
	bc := NewBlockchain()
	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}
