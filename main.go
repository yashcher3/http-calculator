package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func Calc(expression string) (float64, error) {

	if !isValid(expression) {
		return 0, fmt.Errorf("Invalid expression")
	}

	postfix := infixToPostfix(expression)

	result, err := evaluatePostfix(postfix)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func isValid(expression string) bool {
	openCount := strings.Count(expression, "(")
	closeCount := strings.Count(expression, ")")
	if openCount != closeCount {
		return false
	}

	return true
}

func infixToPostfix(expression string) []string {
	var postfix []string
	var stack []string

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	tokens := strings.Split(expression, "")

	for _, token := range tokens {
		if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for stack[len(stack)-1] != "(" {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
		} else if _, isOperator := precedence[token]; isOperator {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else {
			postfix = append(postfix, token)
		}
	}

	for len(stack) > 0 {
		postfix = append(postfix, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return postfix
}

func evaluatePostfix(postfix []string) (float64, error) {
	var stack []float64

	for _, token := range postfix {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return 0, fmt.Errorf("Invalid expression")
			}

			num2 := stack[len(stack)-1]
			num1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			switch token {
			case "+":
				stack = append(stack, num1+num2)
			case "-":
				stack = append(stack, num1-num2)
			case "*":
				stack = append(stack, num1*num2)
			case "/":
				stack = append(stack, num1/num2)
			}
		}
	}

	if len(stack) != 1 {
		return 0, fmt.Errorf("Invalid expression")
	}

	return stack[0], nil
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		body, err_body := io.ReadAll(r.Body)

		defer r.Body.Close()

		jsonData := body
		type Exp struct {
			Expression string
		}
		var exp Exp
		error_json_reading := json.Unmarshal([]byte(jsonData), &exp)
		if error_json_reading != nil {
			fmt.Print(error_json_reading)
		}
		answ, er_calculation := Calc(exp.Expression)

		if er_calculation == nil && err_body == nil && error_json_reading == nil {
			fmt.Fprintf(w, "{\"result\" : \"%s\"}", fmt.Sprint(answ))
		} else if err_body != nil || error_json_reading != nil {
			http.Error(w, "{\"error\" : \"Internal server error\"}", http.StatusInternalServerError)
		} else {
			http.Error(w, "{\"error\" : \"Expression is not valid\"}", http.StatusUnprocessableEntity)
		}
	} else {
		http.Error(w, "{\"error\" : \"Internal server error\"}", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/api/v1/calculate", handler)
	fmt.Println("Сервер запущен на порту 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
	}
}
