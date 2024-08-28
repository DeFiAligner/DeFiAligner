package pathfeat

import (
	"os"
	"strings"
)

//  Reduce symbols to shorter lengths

func readFileToString(filePath string) (string, error) {
	// Read the entire file into a byte slice
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Convert the byte slice to a string
	return string(content), nil
}

type symbolInfo struct {
	symbol string
	value  []string
}

func extractElements(symbol_string string) ([]symbolInfo, string, []string) {

	lines := strings.Split(symbol_string, "\n")
	var symbol_info_list []symbolInfo
	first_line := lines[0]
	line_num := len(lines)
	symbol_num1 := 0
	symbol_num2 := 0

	for len(first_line) > 0 {

		finished_line := 1*symbol_num1 + 2*symbol_num2

		contains_bool_check := strings.Contains(strings.ReplaceAll(lines[line_num-finished_line-1], ")", ""), "#x0000000000000000000000000000000000000000000000000000000000000001") || strings.Contains(strings.ReplaceAll(lines[line_num-finished_line-1], ")", ""), "#x0000000000000000000000000000000000000000000000000000000000000000")

		if strings.HasPrefix(first_line, "(=") && contains_bool_check {
			first_line = first_line[3:]
			symbol_info_list = append(symbol_info_list, symbolInfo{symbol: "(=", value: []string{strings.ReplaceAll(lines[line_num-finished_line-1], ")", "")}})
			symbol_num1 += 1
		} else if strings.HasPrefix(first_line, "(ite") {
			first_line = first_line[5:]

			var string_list []string
			string_list = append(string_list, strings.ReplaceAll(lines[line_num-finished_line-2], ")", ""))
			string_list = append(string_list, strings.ReplaceAll(lines[line_num-finished_line-1], ")", ""))
			symbol_info_list = append(symbol_info_list, symbolInfo{symbol: "(ite", value: string_list})

			symbol_num2 += 1

		} else {
			break
		}
	}

	finished_line := 1*symbol_num1 + 2*symbol_num2
	left_lines := lines[1 : line_num-finished_line]

	return symbol_info_list, first_line, left_lines

}

func SimplifySymbolLogic(symbol_string string) string {

	logic_symbol, first_line, left_lines := extractElements(symbol_string)

	condition_value := "one"
	// Recursive judgment logic
	for _, symbol_info := range logic_symbol {
		if symbol_info.symbol == "(=" {
			contains01 := strings.Contains(symbol_info.value[0], "01")
			switch {
			case contains01 && condition_value == "one":
				condition_value = "one"
			case !contains01 && condition_value == "one":
				condition_value = "zero"
			case contains01 && condition_value == "zero":
				condition_value = "zero"
			case !contains01 && condition_value == "zero":
				condition_value = "one"
			}

		}
		if symbol_info.symbol == "(ite" {
			contains01 := strings.Contains(symbol_info.value[0], "01")
			switch {
			case contains01 && condition_value == "one":
				// condition_value remains "one"
			case !contains01 && condition_value == "one":
				condition_value = "zero"
			case contains01 && condition_value == "zero":
				// condition_value remains "zero"
			case !contains01 && condition_value == "zero":
				condition_value = "one"
			}

		}

	}

	//  merge logic
	final_assertion := "(= " + first_line + "\n"
	for _, line := range left_lines {
		final_assertion += line
	}

	if condition_value == "one" {
		final_assertion += "\n   TRUE)"
	} else {
		final_assertion += "\n   FALSE)"
	}

	return final_assertion

}
