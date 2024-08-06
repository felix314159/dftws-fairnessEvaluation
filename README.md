# DFTWS Fairness Evaluation
This repo contains two parts of code:
* Part 1: [Golang Code](https://github.com/felix314159/dftws-fairnessEvaluation/blob/main/main.go) to simulate DFTWS winner selection in lots of different scenarios.
* Part 2: Python Code to analyze and plot the resulting data (very simple analysis for now).

Note: DFTWS is used as winner selection algorithm in [gophy](https://github.com/felix314159/gophy).

## How to run the Golang code
* ``` git clone https://github.com/felix314159/dftws-fairnessEvaluation.git ```
* ``` cd dftws-fairnessEvaluation ```
* ``` go mod tidy ```
* ``` go run . ```

## How to run the Python code
Install the dependencies pandas, matplotlib and scipy. Then run the code.
