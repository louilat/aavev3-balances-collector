package balances

import (
	"aavev3-raw-balances-collector/internal/aavetoken"
	"aavev3-raw-balances-collector/internal/erctwenty"
	"aavev3-raw-balances-collector/internal/oracle"
	"aavev3-raw-balances-collector/internal/pool"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type AaveToken struct {
	Name                    string
	UnderlyingAsset         string
	Decimals                uint8
	aTokenContract          *aavetoken.Aavetoken
	vTokenContract          *aavetoken.Aavetoken
	UnderlyingTokenPriceUSD *big.Int
	TreasuryAmount          *big.Int
}

func CollectAaveTokens(client *ethclient.Client, pool *pool.Pool, blockNumber *big.Int) ([]AaveToken, error) {
	allAaveTokens := make([]AaveToken, 0)
	reservesList, err := pool.GetReservesList(&bind.CallOpts{BlockNumber: blockNumber})
	if err != nil {
		return make([]AaveToken, 0), err
	}
	fmt.Printf("   Found %v reserves\n", len(reservesList))

	for i, reserve := range reservesList {
		fmt.Printf("   --> [%v / %v] Treating reserve %v", i+1, len(reservesList), reserve)

		// aToken and vToken contracts + Name and decimals
		aTokenAddress, err := pool.GetReserveAToken(&bind.CallOpts{}, reserve)
		if err != nil {
			return make([]AaveToken, 0), err
		}
		aTokenCtr, err := aavetoken.NewAavetoken(aTokenAddress, client)
		if err != nil {
			return make([]AaveToken, 0), err
		}
		vTokenAddress, err := pool.GetReserveVariableDebtToken(&bind.CallOpts{}, reserve)
		if err != nil {
			return make([]AaveToken, 0), err
		}
		vTokenCtr, err := aavetoken.NewAavetoken(vTokenAddress, client)
		if err != nil {
			return make([]AaveToken, 0), err
		}

		name, err := aTokenCtr.Name(&bind.CallOpts{BlockNumber: blockNumber})
		if err != nil {
			return make([]AaveToken, 0), err
		}
		decimals, err := aTokenCtr.Decimals(&bind.CallOpts{BlockNumber: blockNumber})
		if err != nil {
			return make([]AaveToken, 0), err
		}

		// Treasury
		underlyingCtr, err := erctwenty.NewErctwenty(reserve, client)
		if err != nil {
			return make([]AaveToken, 0), err
		}
		treasury, err := underlyingCtr.BalanceOf(&bind.CallOpts{BlockNumber: blockNumber}, common.HexToAddress("0x464C71f6c2F760DdA6093dCB91C24c39e5d6e18c"))
		if err != nil {
			return make([]AaveToken, 0), err
		}
		// Underlying token price in USD
		oracleCtr, err := oracle.NewOracle(common.HexToAddress("0x54586bE62E3c3580375aE3723C145253060Ca0C2"), client)
		if err != nil {
			return make([]AaveToken, 0), err
		}
		tokenPriceUSD, err := oracleCtr.GetAssetPrice(&bind.CallOpts{BlockNumber: blockNumber}, reserve)
		if err != nil {
			return make([]AaveToken, 0), err
		}
		allAaveTokens = append(allAaveTokens, AaveToken{
			Name:                    name,
			UnderlyingAsset:         reserve.String(),
			Decimals:                decimals,
			aTokenContract:          aTokenCtr,
			vTokenContract:          vTokenCtr,
			UnderlyingTokenPriceUSD: new(big.Int).Set(tokenPriceUSD),
			TreasuryAmount:          new(big.Int).Set(treasury),
		})
		fmt.Println("   Done!")
	}
	return allAaveTokens, nil
}
