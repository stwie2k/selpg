package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"math"
)

type Selpg struct {
	Begin int
	End int
	/* false for static line number */
	PageType bool
	Length int
	Destination string
	Src string
	data []string
	Logfile *os.File
}

func process_args(args *Selpg){
	flag.IntVar(&(args.Begin), "s", -1, "start page")
	flag.IntVar(&(args.End), "e", -1, "end page")
	flag.IntVar(&(args.Length), "l", -1, "page len")
	flag.StringVar(&(args.Destination), "d", "", "print destionation")
	flag.BoolVar(&(args.PageType), "f", false, "type of print")
	
	flag.Parse()

	// 处理输入错误
	/*开始结束页错误*/
	
	if args.Begin < 1 || args.Begin > (math.MaxInt32-1) {
		os.Stderr.Write([]byte("Error: Invalid start page\n"))
		os.Exit(1)
	}
	if args.End < 1 || args.End > (math.MaxInt32-1) || args.End < args.Begin {
		os.Stderr.Write([]byte("Error: Invalid end page\n"))
		os.Exit(2)
	}
	/*同时存在-f和-l参数*/
	if args.PageType != false && args.Length != -1 {
		fmt.Fprintln(os.Stderr, "Error: Conflict flags: -f and -l")
		os.Exit(0)
	}
	/*设置行数默认值*/
	if args.PageType== false && args.Length <=0 {
		fmt.Println("Use 72 lines per page as default.")
		args.Length = 72
	}

	if len(flag.Args()) == 1 {
		args.Src = flag.Args()[0]
	} else if len(flag.Args()) > 1 {
		fmt.Fprintf(os.Stderr, "Error: Too much argument. Use selpg -help to know more.\n")
		os.Exit(0)
	} else {
		args.Src = ""
	}
}

func main() {
	// ----定义flag
	flag.Usage = func() {
		fmt.Printf("Usage of seplg:\n")
		fmt.Printf("seplg -s num1 -e num2 [-f -l num3 -d str1 file]\n")
		flag.PrintDefaults()
	}

	// ----实例化对象
	data := new (Selpg)

	process_args(data)



	// ----运行
	// 因为我不知道类似java的切片怎么去用，所以只能这种很丑的代码去完成log操作
	data.Read()
	if data.Destination == "" {
		data.Write()
	} else {
		data.Print()
	}

}

func (selpg *Selpg) Read() {
	if selpg == nil {
		fmt.Fprintf(os.Stderr, "Error: Unknown error.\n")
		
		os.Exit(0)
	}
	var in io.Reader

	// ----确定内容来源
	if selpg.Src == "" {
		in = os.Stdin
	} else {
		var err error
		in, err = os.Open(selpg.Src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: No such file found. Please pass right path.\n")
			
			os.Exit(0)
		}
	}
	
	// ----读取内容
	scanner := bufio.NewScanner(in)
	if selpg.PageType == false { 
		
		count := 0
		for scanner.Scan() {
			line := scanner.Text()
			if count / selpg.Length + 1 >= selpg.Begin && count / selpg.Length + 1 <= selpg.End {
				selpg.data = append(selpg.data, line)
			}
			count++
		}
	} else {
		//-f类型
		cnt := 1
		onComma := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			for i := 0; i < len(data); i++ {
				if data[i] == '\f' {
					return i + 1, data[:i], nil
				}
			}
			if atEOF {
				return 0, data, bufio.ErrFinalToken
			} else {
				return 0, nil, nil
			}
		}
		scanner.Split(onComma)
		for scanner.Scan() {
			line := scanner.Text()
			if cnt >= selpg.Begin && cnt <= selpg.End {
				selpg.data = append(selpg.data, line)
			}
			cnt++
		}
	}
	
}

// ----输出内容到stdout
func (selpg *Selpg) Write() {
	if selpg == nil {
		fmt.Fprintf(os.Stderr, "Error: Unknown error.\n")

		os.Exit(0)
	}
	for i := 0; i < len(selpg.data); i++ {
		fmt.Fprintln(os.Stdout, selpg.data[i])
	}
}

// ----连接到打印机
func (selpg *Selpg) Print() {
	if selpg.Destination != "" {
		file, err := os.Create(selpg.Destination)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Can not create such file")
	
			os.Exit(0)
		}
		for i := 0; i < len(selpg.data); i++ {
			file.WriteString(selpg.data[i])
			file.WriteString("\n")
		}
	}
}