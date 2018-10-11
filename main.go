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
	Infile string
	data []string
}

func process_args(args *Selpg){
	flag.IntVar(&(args.Begin), "s", -1, "the start page")
	flag.IntVar(&(args.End), "e", -1, "the end page")
	flag.IntVar(&(args.Length), "l", -1, "page length")
	flag.StringVar(&(args.Destination), "d", "", "print destionation")
	flag.BoolVar(&(args.PageType), "f", false, "type of print")
	
	flag.Parse()

	// 处理输入错误
	/*开始结束页错误*/
	
	if args.Begin < 1 || args.Begin > (math.MaxInt32-1) {
		os.Stderr.Write([]byte("Invalid start page\n"))
		os.Exit(1)
	}
	if args.End < 1 || args.End > (math.MaxInt32-1) || args.End < args.Begin {
		os.Stderr.Write([]byte("Invalid end page\n"))
		os.Exit(2)
	}
	/*同时存在-f和-l参数*/
	if args.PageType != false && args.Length != -1 {
		fmt.Fprintln(os.Stderr, "Conflict flags -f and -l")
		os.Exit(0)
	}
	/*设置行数默认值*/
	if args.PageType== false && args.Length <=0 {
		fmt.Println("Use 72 lines per page as default.")
		args.Length = 72
	}
    
	if len(flag.Args()) == 1 {
		args.Infile = flag.Args()[0]
	}  else {
		args.Infile = ""
	}
}
func usage(){
	// 定义flag

	flag.Usage = func() {
		fmt.Printf("Usage of seplg:\n")
		fmt.Printf("seplg -s num1 -e num2 [-f -l num3 -d str1 file]\n")
		flag.PrintDefaults()
	}
}
func main() {
	usage()

	// 实例化对象
	data := new (Selpg)

	process_args(data)
    process_input(data)
	
	if data.Destination == "" {
		data.outputrouter()
	} else {
		data.outprint()
	}

}

func  process_input(selpg *Selpg) {
	
	var in io.Reader


	if selpg.Infile == "" {
		in = os.Stdin
	} else {
		var err error
		in, err = os.Open(selpg.Infile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open input file \"%s\"\n", selpg.Infile)
			
			os.Exit(0)
		}
	}

	scanner := bufio.NewScanner(in)
	if ! selpg.PageType { 
		
		count := 0
		for scanner.Scan() {
			line := scanner.Text()
			if count / selpg.Length + 1 >= selpg.Begin && count / selpg.Length  < selpg.End {
				selpg.data = append(selpg.data, line)
			}
			count++
		}
	} else {
		//-f类型
		count := 1
		onSp := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
		scanner.Split(onSp)
		for scanner.Scan() {
			line := scanner.Text()
			if count >= selpg.Begin && count <= selpg.End {
				selpg.data = append(selpg.data, line)
			}
			count++
		}
	}
	
}

// ----输出内容到stdout
func (selpg *Selpg) outputrouter() {



 for i := 0; i < len(selpg.data); i++ {
		fmt.Fprintln(os.Stdout, selpg.data[i])
	}
}

// 连接到打印机


func (selpg *Selpg) outprint() {
	if selpg.Destination != "" {
		file, err := os.Create(selpg.Destination)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Can not create such file")
	
			os.Exit(0)
		}
		for i := 0; i < len(selpg.data); i++ {
			file.WriteString(selpg.data[i])
			file.WriteString("\n")
		}
	}
}