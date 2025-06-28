package balances

import (
	"aavev3-raw-balances-collector/internal/dataprovider"
	"aavev3-raw-balances-collector/internal/pool"
	"aavev3-raw-balances-collector/internal/prevdataprovider"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
)

type UserTokenBalance struct {
	User                    string   `json:"user"`
	BlockNumber             *big.Int `json:"blockNumber"`
	TokenName               string   `json:"tokenName"`
	UnderlyingAsset         string   `json:"underlyingAsset"`
	Decimals                uint8    `json:"decimals"`
	ScaledATokenBalance     *big.Int `json:"scaledATokenBalance"`
	ScaledVariableDebt      *big.Int `json:"scaledVariableDebt"`
	UnderlyingTokenPriceUSD *big.Int `json:"underlyingTokenPriceUSD"`

	TotalCollateralBase         *big.Int `json:"totalCollateralBase"`
	TotalDebtBase               *big.Int `json:"totalDebtBase"`
	CurrentLiquidationThreshold *big.Int `json:"currentLiquidationThreshold"`
	Ltv                         *big.Int `json:"ltv"`
	HealthFactor                *big.Int `json:"healthFactor"`
	UserEMode                   *big.Int `json:"userEMode"`
}

func CollectUserBalancesManual(user string, pool *pool.Pool, tokens []AaveToken, blockNumber *big.Int) ([]UserTokenBalance, error) {
	fmt.Printf("   --> Collecting balance of user %v\n", user)
	userAddress := common.HexToAddress(user)
	userTokenBalances := make([]UserTokenBalance, 0)

	userData, err := pool.GetUserAccountData(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	userEMode, err := pool.GetUserEMode(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	fmt.Printf("   ")
	for _, token := range tokens {
		fmt.Printf(".")
		a, err := token.aTokenContract.ScaledBalanceOf(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
		if err != nil {
			return make([]UserTokenBalance, 0), err
		}
		v, err := token.vTokenContract.ScaledBalanceOf(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
		if err != nil {
			return make([]UserTokenBalance, 0), err
		}

		if a.Cmp(common.Big0) > 0 || v.Cmp(common.Big0) > 0 {
			userTokenBalances = append(userTokenBalances, UserTokenBalance{
				User:                        user,
				BlockNumber:                 blockNumber,
				TokenName:                   token.Name,
				UnderlyingAsset:             token.UnderlyingAsset,
				Decimals:                    token.Decimals,
				ScaledATokenBalance:         new(big.Int).Set(a),
				ScaledVariableDebt:          new(big.Int).Set(v),
				UnderlyingTokenPriceUSD:     token.UnderlyingTokenPriceUSD,
				TotalCollateralBase:         userData.TotalCollateralBase,
				TotalDebtBase:               userData.TotalDebtBase,
				CurrentLiquidationThreshold: userData.CurrentLiquidationThreshold,
				Ltv:                         userData.Ltv,
				HealthFactor:                userData.HealthFactor,
				UserEMode:                   userEMode,
			})
		}
	}
	fmt.Println(" Done!")
	return userTokenBalances, nil
}

func CollectUserBalancesPrev(user string, pool *pool.Pool, dataProvider *prevdataprovider.Prevdataprovider, tokens []AaveToken, blockNumber *big.Int) ([]UserTokenBalance, error) {
	fmt.Printf("   --> Collecting balance of user %v\n", user)
	userAddress := common.HexToAddress(user)
	userTokenBalances := make([]UserTokenBalance, 0)

	userData, err := pool.GetUserAccountData(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	userEMode, err := pool.GetUserEMode(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	poolAddressesProvider := common.HexToAddress("0x2f39d218133AFaB8F2B819B1066c7E434Ad94E9e")
	userBalances, _, err := dataProvider.GetUserReservesData(&bind.CallOpts{BlockNumber: blockNumber}, poolAddressesProvider, common.Address(userAddress.Bytes()))
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	for _, balance := range userBalances {
		if balance.ScaledATokenBalance.Cmp(common.Big0) > 0 || balance.ScaledVariableDebt.Cmp(common.Big0) > 0 {
			// Find corresponding token
			var token AaveToken
			for _, tkn := range tokens {
				if tkn.UnderlyingAsset == balance.UnderlyingAsset.String() {
					token = tkn
				}
			}
			userTokenBalances = append(userTokenBalances, UserTokenBalance{
				User:                        user,
				BlockNumber:                 blockNumber,
				TokenName:                   token.Name,
				UnderlyingAsset:             balance.UnderlyingAsset.String(),
				Decimals:                    token.Decimals,
				ScaledATokenBalance:         new(big.Int).Set(balance.ScaledATokenBalance),
				ScaledVariableDebt:          new(big.Int).Set(balance.ScaledVariableDebt),
				UnderlyingTokenPriceUSD:     token.UnderlyingTokenPriceUSD,
				TotalCollateralBase:         userData.TotalCollateralBase,
				TotalDebtBase:               userData.TotalDebtBase,
				CurrentLiquidationThreshold: userData.CurrentLiquidationThreshold,
				Ltv:                         userData.Ltv,
				HealthFactor:                userData.HealthFactor,
				UserEMode:                   userEMode,
			})
		}
	}
	fmt.Println(" Done!")
	return userTokenBalances, nil
}

func CollectUserBalances(user string, pool *pool.Pool, dataProvider *dataprovider.Dataprovider, tokens []AaveToken, blockNumber *big.Int) ([]UserTokenBalance, error) {
	fmt.Printf("   --> Collecting balance of user %v\n", user)
	userAddress := common.HexToAddress(user)
	userTokenBalances := make([]UserTokenBalance, 0)

	userData, err := pool.GetUserAccountData(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	userEMode, err := pool.GetUserEMode(&bind.CallOpts{BlockNumber: blockNumber}, userAddress)
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	poolAddressesProvider := common.HexToAddress("0x2f39d218133AFaB8F2B819B1066c7E434Ad94E9e")
	userBalances, _, err := dataProvider.GetUserReservesData(&bind.CallOpts{BlockNumber: blockNumber}, poolAddressesProvider, common.Address(userAddress.Bytes()))
	if err != nil {
		return make([]UserTokenBalance, 0), err
	}

	for _, balance := range userBalances {
		if balance.ScaledATokenBalance.Cmp(common.Big0) > 0 || balance.ScaledVariableDebt.Cmp(common.Big0) > 0 {
			// Find corresponding token
			var token AaveToken
			for _, tkn := range tokens {
				if tkn.UnderlyingAsset == balance.UnderlyingAsset.String() {
					token = tkn
				}
			}
			userTokenBalances = append(userTokenBalances, UserTokenBalance{
				User:                        user,
				BlockNumber:                 blockNumber,
				TokenName:                   token.Name,
				UnderlyingAsset:             balance.UnderlyingAsset.String(),
				Decimals:                    token.Decimals,
				ScaledATokenBalance:         new(big.Int).Set(balance.ScaledATokenBalance),
				ScaledVariableDebt:          new(big.Int).Set(balance.ScaledVariableDebt),
				UnderlyingTokenPriceUSD:     token.UnderlyingTokenPriceUSD,
				TotalCollateralBase:         userData.TotalCollateralBase,
				TotalDebtBase:               userData.TotalDebtBase,
				CurrentLiquidationThreshold: userData.CurrentLiquidationThreshold,
				Ltv:                         userData.Ltv,
				HealthFactor:                userData.HealthFactor,
				UserEMode:                   userEMode,
			})
		}
	}
	fmt.Println(" Done!")
	return userTokenBalances, nil
}

func CollectAllUsersBalancesManual(users []string, pool *pool.Pool, tokens []AaveToken, blockNumber *big.Int) ([]UserTokenBalance, error) {
	allUsersBalances := make([]UserTokenBalance, 0)

	for idx, user := range users {
		fmt.Printf("[%v / %v]\n", idx, len(users))
		userBalances, err := CollectUserBalancesManual(user, pool, tokens, blockNumber)
		if err != nil {
			return make([]UserTokenBalance, 0), err
		}
		allUsersBalances = append(allUsersBalances, userBalances...)
	}
	return allUsersBalances, nil
}

func CollectAllUsersBalancesPrev(users []string, pool *pool.Pool, dataProvider *prevdataprovider.Prevdataprovider, tokens []AaveToken, blockNumber *big.Int) ([]UserTokenBalance, error) {
	allUsersBalances := make([]UserTokenBalance, 0)

	for idx, user := range users {
		fmt.Printf("[%v / %v]\n", idx, len(users))
		userBalances, err := CollectUserBalancesPrev(user, pool, dataProvider, tokens, blockNumber)
		if err != nil {
			return make([]UserTokenBalance, 0), err
		}
		allUsersBalances = append(allUsersBalances, userBalances...)
	}
	return allUsersBalances, nil
}

func CollectAllUsersBalances(users []string, pool *pool.Pool, dataProvider *dataprovider.Dataprovider, tokens []AaveToken, blockNumber *big.Int) ([]UserTokenBalance, error) {
	allUsersBalances := make([]UserTokenBalance, 0)

	for idx, user := range users {
		fmt.Printf("[%v / %v]\n", idx, len(users))
		userBalances, err := CollectUserBalances(user, pool, dataProvider, tokens, blockNumber)
		if err != nil {
			return make([]UserTokenBalance, 0), err
		}
		allUsersBalances = append(allUsersBalances, userBalances...)
	}
	return allUsersBalances, nil
}
