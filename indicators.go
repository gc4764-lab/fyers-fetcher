package main

import (
	"fmt"
)

// CalculateSMA computes the Simple Moving Average for a given period.
// It returns a slice of the same length as the input, with leading zeros where the SMA cannot be calculated.
func CalculateSMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}
	
	sma := make([]float64, len(prices))
	var currentSum float64

	// Calculate the sum for the first window
	for i := 0; i < period; i++ {
		currentSum += prices[i]
	}
	sma[period-1] = currentSum / float64(period)

	// Slide the window for the remaining prices
	for i := period; i < len(prices); i++ {
		currentSum = currentSum - prices[i-period] + prices[i]
		sma[i] = currentSum / float64(period)
	}
	
	return sma
}

// CalculateRSI computes the Relative Strength Index using Wilder's Smoothing.
func CalculateRSI(prices []float64, period int) []float64 {
	if len(prices) <= period {
		return nil
	}

	rsi := make([]float64, len(prices))
	var avgGain, avgLoss float64

	// Step 1: Calculate the initial Average Gain and Average Loss (Simple Average)
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			avgGain += change
		} else {
			avgLoss -= change // keep losses as positive numbers
		}
	}
	
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate the first RSI value
	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - (100 / (1 + rs))
	}

	// Step 2: Calculate subsequent RSI values using Wilder's Smoothing
	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		var currentGain, currentLoss float64
		
		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}

		// Wilder's Smoothing technique
		avgGain = ((avgGain * float64(period-1)) + currentGain) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + currentLoss) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi
}

func main() {
	// Mock daily closing prices for NSE:RELIANCE-EQ
	closingPrices := []float64{
		2910.5, 2925.0, 2915.2, 2890.0, 2875.5,
		2880.0, 2905.0, 2930.5, 2950.0, 2965.2,
		2980.0, 2975.5, 2990.0, 3010.5, 3025.0,
		3015.0, 3000.5, 2985.0, 2995.5, 3020.0,
	}

	smaPeriod := 5
	rsiPeriod := 14

	smaData := CalculateSMA(closingPrices, smaPeriod)
	rsiData := CalculateRSI(closingPrices, rsiPeriod)

	fmt.Println("Day | Close Price | 5-Day SMA | 14-Day RSI")
	fmt.Println("-------------------------------------------")
	
	for i := 0; i < len(closingPrices); i++ {
		smaVal := "N/A  "
		if smaData[i] > 0 {
			smaVal = fmt.Sprintf("%.2f", smaData[i])
		}
		
		rsiVal := "N/A  "
		if rsiData[i] > 0 {
			rsiVal = fmt.Sprintf("%.2f", rsiData[i])
		}

		fmt.Printf("%2d  | %11.2f | %9s | %s\n", i+1, closingPrices[i], smaVal, rsiVal)
	}
}

