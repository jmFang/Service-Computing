package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	//"flag"
)

/* INBUFSIZ is size of array inbuf */
const INBUFSIZ int = 16 * 1024
const BUFSIZ int = 128

/*================================= types =========================*/
type selpg_page struct {
	start_page  int
	end_page    int
	in_filename string
	page_len    int
	page_type   int
	print_dest  string
}

/*================================= globals =======================*/
var progname string /* program name, for error messages */

/*================================= prototypes ====================*/
func usage() {
	fmt.Printf("\nUSAGE: %s -sstart_page -eend_page [-f] [-llines_per_page ] \n [-ddest ] [ in_filename ]\n", progname)
}

func process_args(ac int, av []string, psa *selpg_page) {
	var s1 string
	var s2 string
	var argno int

	var i int

	if ac < 3 {
		fmt.Printf("%s: not enough arguments\n", progname)
		usage()
		os.Exit(1)
	}

	s1 = av[1]

	fmt.Println(s1[2] - 48)

	if strings.Index(s1, "-s") != 0 {
		fmt.Printf("%s: 1st arg should be -sstart_page\n", progname)
		usage()
		os.Exit(2)
	}

	i = int(s1[2] - 48)

	if i < 1 {
		fmt.Printf("%s : invalid start page %s\n", progname, string(s1[2]))
		usage()
		os.Exit(3)
	}

	psa.start_page = i

	s1 = av[2]

	if strings.Index(s1, "-e") != 0 {
		fmt.Printf("%s: 2nd arg should -eend_page\n", progname)
		usage()
		os.Exit(4)
	}

	i = int(s1[2] - 48)

	if i < 1 || i < psa.start_page {
		fmt.Printf("%s: invalid end page %s\n", progname, string(s1[2]))
		usage()
		os.Exit(5)
	}

	psa.end_page = i

	argno = 3

	for argno <= ac-1 && strings.Compare(string(av[argno][0]), string('-')) == 0 {
		/* while there more args and they start with a '-' */
		s1 = av[argno]

		switch s1[1] {

		case 'l':
			s2 = strings.TrimPrefix(s1, "-l")
			//string to int64
			k, _ := strconv.ParseInt(s2, 10, 64)
			if k < 1 {
				fmt.Printf("%s: invalid page length %s\n", progname, s2)
				usage()
				os.Exit(6)
			}

			psa.page_len = int(k) //强制换回int，一般不会出问题

			argno++
			continue

		case 'f':
			if strings.Compare(s1, "-f") != 0 {
				fmt.Printf("%s: option should be \"-f\"\n", progname)
				usage()
				os.Exit(7)
			}

			psa.page_type = 'f'
			argno++
			continue

		case 'd':
			s2 = strings.TrimPrefix(s1, "-d")
			if len(s2) < 1 {
				fmt.Printf("%s: -d option requires a printer destination", progname)
				usage()
				os.Exit(8)
			}

			psa.print_dest = s2
			argno++
			continue

		default:
			fmt.Printf("%s: unknown option %s\n", progname, s1)
			usage()
			os.Exit(9)

		} /*end switch*/
	} /*end while*/

	if argno <= ac-1 {

		psa.in_filename = av[argno]

		fmt.Printf("%s reading file\n", psa.in_filename)

		if _, err := os.Open(psa.in_filename); err != nil {
			fmt.Printf("%s: input file \"%s\" does not exist\n", progname, psa.in_filename)
			fmt.Sprintln(err, "%s: input file \"%s\" does not exist\n", progname, psa.in_filename)
			os.Exit(10)
		}

		if _, err := os.OpenFile(psa.in_filename, os.O_RDWR|os.O_CREATE, 0755); err != nil {
			fmt.Printf("%s: input file \"%s\" exists but cannot be read\n", progname, psa.in_filename)
			os.Exit(11)
		}
	}

	/* check some post-conditions */
	//assert(psa->start_page > 0);
	//assert(psa->end_page > 0 && psa->end_page >= psa->start_page);
	//assert(psa->page_len > 1);
	//assert(psa->page_type == 'l' || psa->page_type == 'f');

}

func process_input(sa selpg_page) {
	var fin *os.File
	/* input stream */
	var fout *os.File /* output stream */
	//var temp = make([]byte, BUFSIZ)   /* temp string var */
	//var crc string  /* for char ptr return code */
	//var c int /* to read 1 char */
	var line = make([]byte, BUFSIZ)
	var line_ctr int /* line counter */
	var page_ctr int /* page counter */
	//var inbuf = make([]byte, INBUFSIZ) /* for better performance on input stream */

	/* set the input source */

	if sa.in_filename == "" {
		fin = os.Stdin
	} else {
		fin, _ = os.OpenFile(sa.in_filename, os.O_RDWR|os.O_CREATE, 0755)
		fmt.Printf("fin is reading file : %s\n", sa.in_filename)
		if fin == nil {
			fmt.Printf("%s: could not open input file \"%s\"\n", progname, sa.in_filename)
			os.Exit(12)
		}

	}

	/* set the output destination */

	if sa.print_dest == "" {
		fout = os.Stdout
		fmt.Printf("%s stdout\n", sa.in_filename)
	} else {
		// 刷新输出缓冲区
		fout.Sync()
		//sprintf(s1, "lp -d%s", sa.print_dest);
		str := fmt.Sprintf("-d%s", sa.print_dest)

		cmd := exec.Command("lp", str)

		_, err := cmd.Output()

		if err != nil {
			fmt.Printf("%s: could not open pipe to \"%s\"\n", progname, str)
			os.Exit(13)
		}

		cmd.Start()
	}

	/* begin one of two main loops based on page type */

	if sa.page_type == 'l' {
		line_ctr = 0
		page_ctr = 1

		fmt.Println("here")
		fout, _ = os.Create("output.txt")

		rd := bufio.NewReader(fin)
		wd := bufio.NewWriter(fout)

		var err error
		for true {
			line, err = rd.ReadSlice('\n')

			if err != nil || io.EOF == err {
				break
			}

			line_ctr++

			if line_ctr > sa.page_len {
				page_ctr++
				line_ctr = 1
			}
			if page_ctr >= sa.start_page && page_ctr <= sa.end_page {

				_, err := wd.Write(line)

				if err != nil {
					fmt.Printf("%s: error while writing", progname)
					os.Exit(14)
				}
			}
		}
		wd.Flush()
		os.Exit(0)
	} else { //假定页由换页符定界
		page_ctr = 1
		//char := make([]byte, 1)

		fout, _ = os.Create(sa.print_dest)

		rd := bufio.NewReader(fin)
		wd := bufio.NewWriter(fout)

		for true {

			//_, err := fin.Read(char)

			//line, err = rd.ReadSlice('\n')
			line, err := rd.ReadString('\n')

			if err == io.EOF || err != nil {
				break
			}
			//好麻烦，处理一下字符比较也这么麻烦
			if len(line) > 0 {
				if page_ctr >= sa.start_page && page_ctr <= sa.end_page {
					wd.WriteString(line)
				} else {
					break
				}
				page_ctr++
			}

			/*else {

				if page_ctr >= sa.start_page && page_ctr <= sa.end_page {
					//err := ioutil.WriteFile("./output.txt", char, 0666)
					_, err := wd.Write(line)
					if err != nil {
						fmt.Printf("%s: error while writing", progname)
						os.Exit(15)
					}
				}

			}*/

		}
		wd.Flush()
		os.Exit(0)

	}
	/* end main loop */
	if page_ctr < sa.start_page {
		fmt.Printf("%s: start_page (%d) greater than total pages (%d),no output written\n", progname, sa.end_page, page_ctr)
		os.Exit(17)
	} else if page_ctr < sa.end_page {
		fmt.Printf("%s: end_page (%d) greater than total pages (%d), less output than expected\n", progname, sa.end_page, page_ctr)
	}

	fin.Close()
	fout.Sync()
	fmt.Printf("%s: done", progname)

}

func main() {

	var sa selpg_page

	av := os.Args
	ac := len(os.Args)

	for _, str := range av {
		fmt.Printf("%s \n", str)
	}
	argCount := len(os.Args[1:])

	progname = os.Args[0]

	sa.start_page = -1
	sa.end_page = -1
	sa.page_len = 72
	sa.in_filename = ""
	sa.print_dest = ""
	sa.page_type = 'l'

	process_args(ac, av, &sa)

	process_input(sa)

	fmt.Printf("Total Arguments (excluding program name): %d\n", argCount)

	return
}
