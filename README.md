# 服务计算作业——用golang实现linux下命令selpg



---
## 内容简介
这次博客的主要内容为使用 golang 开发  Linux 命令行实用程序 中的 selpg。主要的思路就是将c语言中的selpg的写法翻译成go。大部分内容参照了[开发 Linux 命令行实用程序][1]

##  功能说明
selpg是一个unix系统下命令。

该命令本质上就是将一个文件，通过自己设定的分页方式，输出到屏幕或者重定向到其他文件上，或者利用打印机打印出来。使用格式如下。

    -s start_page -e end_page [ -f | -l lines_per_page ][ -d dest ] [ in_filename ]


必要参数

 - -s，后面接开始读取的页号 int
 - -e，后面接结束读取的页号 int s和e都要大于1，并且s <= e，否则提示错误
s和e都要大于1，并且s <= e，否则提示错误
可选参数：

 - -l，后面跟行数 int，代表多少行分为一页，不指定 -l 又缺少 -f  -
则默认按照72行分一页
 - -f，该标志无参数，代表按照分页符’\f’ 分页
 - -d，后面接打印机名称，用于打印filename，唯一一个无标识参数，代表选择读取的文件名



## 设计思路
其实主要是参照了c语言的selpg的写法，相对有些机械的翻译了。

1.设计selpg结构体
给selpg设计一个结构体，分别保存各个参数的值，使其易于处理。

    type Selpg struct {
    	Begin int //起始页
    	End int //结束页
    	PageType bool //是否为-f类型命令，是则为真
    	Length int //页长度
    	Destination string //打印机名称
    	Infile string//输入文件
    	data []string//储存文件数据的字符串
    }


2.利用flag包解析参数
首先定义一个flag。

    flag.Usage = func() {
    		fmt.Printf("Usage of seplg:\n")
    		fmt.Printf("seplg -s num1 -e num2 [-f -l num3 -d str1 file]\n")
    		flag.PrintDefaults()
    	}
然后利用flag参数来解析各个参数

    flag.IntVar(&(args.Begin), "s", -1, "the start page")
    	flag.IntVar(&(args.End), "e", -1, "the end page")
    	flag.IntVar(&(args.Length), "l", -1, "page length")
    	flag.StringVar(&(args.Destination), "d", "", "print destionation")
    	flag.BoolVar(&(args.PageType), "f", false, "type of print")

3.处理分页
分页的时候主要需要应对两种情况。一种是-f以/f为分页的标志。另一种是-l指定分页的行数。使用scanner来读取每一行获取其信息。
-f时搜寻\f符号，利用split函数将文本分隔开，每个作为一页。

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

-l时则计算行数，当行数打到指定行数进行分页。

    for scanner.Scan() {
    			line := scanner.Text()
    			if count / selpg.Length + 1 >= selpg.Begin && count / selpg.Length  < selpg.End {
    				selpg.data = append(selpg.data, line)
    			}
    			count++
    		}


4.重定向输入输出
这里采用的方法是将读入的内容储存到一个string字符串中，等到需要输出时再重定向到标准输出或是打印机上。不过更好的方法是通过管道把selpg输出给另一个命令。方法是调用函数 exec.Command("命令名", "该命令的相关参数", ...) 即可返回获得该命令结构，再调用 cmd.StdinPipe() 可以获得管道的输入管道，然后把我们的 selpg 输出到该管道即可。 这里给出范例。

    if sa.printDest != "" {
        cmd := exec.Command("grep", "-nf", "keyword")
        inpipe, err = cmd.StdinPipe()
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        defer inpipe.Close()
        cmd.Stdout = fout
        cmd.Start()
    }








## 测试结果

### 1、输入输出重定向范例

 `selpg -s=1 -e=1 a.txt`

该命令将把a.txt的第1页写至标准输出（也就是屏幕）。
![此处输入图片的描述][2]

    selpg -s=1 -e=1 < a.txt
该命令效果同上，在本例中，selpg 读取标准输入，而标准输入已被 shell／内核重定向为来自“input_file”而不是显式命名的文件名参数。输入的第 1 页被写至屏幕。
![此处输入图片的描述][3]

 
    selpg -s=1 -e=1 a.txt >b.txt 

将a.txt的第 1 页写至标准输出；标准输出被 shell／内核重定向至“b.txt"。即将a.txt中内容输入至b.txt中。

打开b.txt
![此处输入图片的描述][4]
重定向成功

### 2、-l类型指令

    selpg -s=1 -e=1 -l=3 a.txt  

从a.txt输入，并输出到命令行，打印前三行
![此处输入图片的描述][5]


    selpg -s=1 -e=2 -l=3 a.txt  
从a.txt输入，并输出到命令行，打印两页，总共六行
![此处输入图片的描述][6]

    selpg -s=2 -e=2  -l=3 a.txt  
打印出a.txt的第二页

![此处输入图片的描述][7]

### 3、-f类型指令
    selpg -s=1 -e=1 -f a.txt
根据换页符-f分页，打出第一页
![此处输入图片的描述][8]

  

     selpg -s=1 -e=3 -f a.txt
根据换页符-f分页，打出前三页
![此处输入图片的描述][9]

     selpg -s=1 -e a.txt 2>error.txt
这里故意输错命令，将错误信息写入error.txt
![此处输入图片的描述][10]
     
打开error.txt可以看到错误信息
![此处输入图片的描述][11]


具体的代码可参照[github][12]


  [1]: https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html
  [2]: https://i.loli.net/2018/10/10/5bbdaf97138d4.png
  [3]: https://i.loli.net/2018/10/10/5bbdb1145b384.png
  [4]: https://i.loli.net/2018/10/11/5bbf28495187e.png
  [5]: https://i.loli.net/2018/10/11/5bbf2a450f022.png
  [6]: https://i.loli.net/2018/10/11/5bbf2aa41bf21.png
  [7]: https://i.loli.net/2018/10/11/5bbf2c81a05c4.png
  [8]: https://i.loli.net/2018/10/11/5bbf2e25502b0.png
  [9]: https://i.loli.net/2018/10/11/5bbf2e734e414.png
  [10]: https://i.loli.net/2018/10/11/5bbf2f2b49a67.png
  [11]: https://i.loli.net/2018/10/11/5bbf2f5932241.png
  [12]: https://github.com/stwie2k/selpg
