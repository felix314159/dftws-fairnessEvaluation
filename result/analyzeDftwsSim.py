# on macos: 								python -m pip install -r requirements.txt
import csv
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
from scipy.stats import chisquare

amountOfRunsPerTrial = 3000
chiSquareSymbol = "$\u03C7^2$"

def main():
	# read simulation results from csv into a list of lists (each row is a list)
	results = readCSV("data.csv")

	# analyze the data
	for trialIndex, trialData in enumerate(results):
		calcAndPlot(trialData, trialIndex+1)
		#print(trialData, "\n\n")


def readCSV(fileName):
	data = []
	with open(fileName, newline='') as csvfile:
		csvreader = csv.reader(csvfile)
        # skip header (first row)
		next(csvreader)
        # get data from all other rows
		for row in csvreader:
			# convert string to int (it is assumed that each element of each row is an int [when you skip the header row])
			intRow = [int(value) for value in row]
			data.append(intRow)
	return data


def calcAndPlot(dataList, trialIndex):
	amountOfNodes = len(dataList)

	# Convert to a pandas DataFrame for convenience
	df = pd.DataFrame({'Node': range(1, amountOfNodes+1), 'Wins': dataList})

	# determine how often each node should have won in a perfectly fair algo
	runsPerTrial = sum(dataList)
	expectedAmountOfWinsPerNode = runsPerTrial / amountOfNodes # 3000 / 40 = 75, this is the arithmetic average
	expectedAmountOfWinsPerNodeList = [expectedAmountOfWinsPerNode] * amountOfNodes

	# calculate chi-square test result
	chi2Result, pValue = chisquare(dataList, f_exp=expectedAmountOfWinsPerNodeList)

	#print(f'Runs per Trial: {runsPerTrial}')
	#print(f'Amount of Node Identities: {amountOfNodes}')
	#print(f'Expected Amount of Wins Per Node: {runsPerTrial}/{amountOfNodes} = {expectedAmountOfWinsPerNode}')
	print(f'Trial #{trialIndex}:')
	print(f'\tLowest Amount of Observed Wins: {min(dataList)}')
	print(f'\tHighest Amount of Observed Wins: {max(dataList)}')
	print(f'\tChi-Square: {chi2Result}')
	print(f'\tP-Value: {pValue}\n---')

	# ---- visualize results ----

	# set windows title and size
	plt.figure(f'Analysis of Trial {trialIndex} (Amount of Runs: {amountOfRunsPerTrial})', figsize=(22, 6))
	plt.bar(df['Node'], df['Wins'], color='blue', alpha=0.6, label='Observed Wins')

	# show expected amount of wins line
	plt.axhline(y=expectedAmountOfWinsPerNode, color='grey', linestyle='--', label='Expected Wins')
	# 		and show the actual value (just from the red line it is hard to tell which exact number it is when you are not looking at cmd output)
	plt.text(amountOfNodes+3, expectedAmountOfWinsPerNode, f'{expectedAmountOfWinsPerNode}', color='grey', va='center') # amountOfNodes+3 just says the x coordinate of the string should be where element 43 would be (so right to the plot), expectedAmountOfWinsPerNode defines the y coordinate of the string (should be same height as the red line), third parameter is the actual value shown as string

	# show observed minimum amount of wins
	plt.axhline(y=min(dataList), color='red', linestyle='--', label='Min observed Wins')
	# 		and show the actual value
	plt.text(amountOfNodes+3, min(dataList), f'{min(dataList)}', color='red', va='center')

	# show observed maximum amount of wins
	plt.axhline(y=max(dataList), color='green', linestyle='--', label='Max observed Wins')
	# 		and show the actual value
	plt.text(amountOfNodes+3, max(dataList), f'{max(dataList)}', color='green', va='center')

	# show calculated chi square value
	roundedChiSquareResult = float('{:,.2f}'.format(chi2Result))
	plt.text(amountOfNodes+3, 26.0, f'{chiSquareSymbol} ≈ {roundedChiSquareResult}', color='purple', va='center')

	# show calculated p-value
	roundedPValueResult = float('{:,.2f}'.format(pValue))
	plt.text(amountOfNodes+3, 32.0, f'p ≈ {roundedPValueResult}', color='brown', va='center')

	plt.xlabel('Node Identifier')
	plt.ylabel('Amount of DFTWS Wins')
	plt.title('Observed vs. Expected Amount Wins per Node')
	plt.legend(bbox_to_anchor=(1, 0.21),ncol=1, fancybox=True, shadow=True) # increase value 0.21 if you want to put the legend higher
	# Set x-axis labels to show 1, 2, 3, ...
	plt.xticks(ticks=df['Node'], labels=df['Node'])
	
	plt.show()


main()

""" Notes:
		Null Hypothesis: 			The winner selection of DFTWS is fair (as in every eligible miner has similar chances to be selected winner)
		P-Value meaning: 			If this value is smaller than 0.05 the null hypothesis should be rejected (as in that would show that DFTWS is not fair). Observed p-value were in 0.13 to 0.97 which means that the performed analysis does not imply that DFTWS is unfair.
		Chi-Square value meaning: 	Its range is [0,inf) and the smaller the value is the fairer DFTWS seems to be (e.g. value 0 would mean observed wins perfectly match expected wins). Observed values were always below 50.
"""