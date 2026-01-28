/**
 * 币种基础配置及工具函数
 */

// 币种预设价格配置（2026年1月数据，用于WebSocket推送前的占位显示）
// 价格来源：CoinMarketCap实时数据，避免显示0价格导致用户体验问题
export const CRYPTO_PRESET_PRICES = {
	// 主流币 (Top 10)
	'BTC': 88800.00,    // Bitcoin
	'ETH': 2960.00,     // Ethereum
	'BNB': 598.00,      // Binance Coin
	'SOL': 127.00,      // Solana
	'XRP': 1.92,        // Ripple
	'ADA': 0.36,        // Cardano
	'DOGE': 0.125,      // Dogecoin
	'TRX': 0.295,       // Tron
	'AVAX': 22.15,      // Avalanche
	'DOT': 4.28,        // Polkadot
	
	// DeFi生态
	'LINK': 13.87,      // Chainlink
	'UNI': 6.85,        // Uniswap
	'AAVE': 242.50,     // Aave
	'CRV': 0.365,       // Curve DAO
	'COMP': 48.50,      // Compound
	'SUSHI': 0.685,     // SushiSwap
	'YFI': 5850.00,     // Yearn Finance
	
	// 老牌公链
	'LTC': 68.95,       // Litecoin
	'BCH': 342.00,      // Bitcoin Cash
	'XMR': 156.30,      // Monero
	'ETC': 18.65,       // Ethereum Classic
	'DASH': 26.80,      // Dash
	'ZEC': 35.20,       // Zcash
	'XTZ': 0.715,       // Tezos
	'NEO': 9.87,        // NEO
	'QTUM': 2.15,       // Qtum
	'ONT': 0.158,       // Ontology
	'ICX': 0.142,       // ICON
	
	// Layer2与扩容
	'MATIC': 0.384,     // Polygon (已更名为POL)
	'ATOM': 4.52,       // Cosmos
	'ALGO': 0.142,      // Algorand
	'FIL': 3.28,        // Filecoin
	'VET': 0.0234,      // VeChain
	'XLM': 0.0912,      // Stellar
	'IOTA': 0.125,      // IOTA
	'THETA': 1.18,      // Theta Network
	'GRT': 0.096,       // The Graph
	'SAND': 0.285,      // The Sandbox
	'MANA': 0.325,      // Decentraland
	
	// 其他项目
	'BAT': 0.175,       // Basic Attention Token
	'ZRX': 0.328,       // 0x Protocol
	'KNC': 0.485,       // Kyber Network
	'ZIL': 0.0145,      // Zilliqa
	'SNX': 1.52,        // Synthetix
	'KSM': 18.50,       // Kusama
	'APE': 1.08,        // ApeCoin
	
	// 平台币/特殊
	'HT': 0.425,        // Huobi Token
	'OMG': 0.285,       // OMG Network
	'MKR': 1285.00      // Maker (未在系统中但保留)
}

// 币种基础配置（用于市场页面展示）
export const cryptoBaseData = [
	// 主流币
	{ name: 'BTC', price: CRYPTO_PRESET_PRICES.BTC, tag: '价格保护', tagClass: 'protect' },
	{ name: 'ETH', price: CRYPTO_PRESET_PRICES.ETH, tag: '价格保护', tagClass: 'protect' },
	{ name: 'BNB', price: CRYPTO_PRESET_PRICES.BNB, tag: '', tagClass: '' },
	{ name: 'SOL', price: CRYPTO_PRESET_PRICES.SOL, tag: '', tagClass: '' },
	{ name: 'XRP', price: CRYPTO_PRESET_PRICES.XRP, tag: '', tagClass: '' },
	{ name: 'ADA', price: CRYPTO_PRESET_PRICES.ADA, tag: '', tagClass: '' },
	{ name: 'DOGE', price: CRYPTO_PRESET_PRICES.DOGE, tag: '', tagClass: '' },
	{ name: 'DOT', price: CRYPTO_PRESET_PRICES.DOT, tag: '', tagClass: '' },
	{ name: 'AVAX', price: CRYPTO_PRESET_PRICES.AVAX, tag: '', tagClass: '' },
	{ name: 'LINK', price: CRYPTO_PRESET_PRICES.LINK, tag: '', tagClass: '' },
	{ name: 'MATIC', price: CRYPTO_PRESET_PRICES.MATIC, tag: '', tagClass: '' },
	{ name: 'LTC', price: CRYPTO_PRESET_PRICES.LTC, tag: '', tagClass: '' },
	{ name: 'BCH', price: CRYPTO_PRESET_PRICES.BCH, tag: '', tagClass: '' },
	{ name: 'XMR', price: CRYPTO_PRESET_PRICES.XMR, tag: '', tagClass: '' },
	{ name: 'ETC', price: CRYPTO_PRESET_PRICES.ETC, tag: '', tagClass: '' },
	{ name: 'ATOM', price: CRYPTO_PRESET_PRICES.ATOM, tag: '', tagClass: '' },
	{ name: 'XLM', price: CRYPTO_PRESET_PRICES.XLM, tag: '', tagClass: '' },
	{ name: 'VET', price: CRYPTO_PRESET_PRICES.VET, tag: '', tagClass: '' },
	{ name: 'TRX', price: CRYPTO_PRESET_PRICES.TRX, tag: '', tagClass: '' },
	{ name: 'NEO', price: CRYPTO_PRESET_PRICES.NEO, tag: '', tagClass: '' },
	{ name: 'EOS', price: 0.5432, tag: '', tagClass: '' },
	// 新公链/Layer2
	{ name: 'APT', price: 5.68, tag: '新币', tagClass: 'new' },
	{ name: 'SUI', price: 1.4863, tag: '新币', tagClass: 'new' },
	{ name: 'ARB', price: 0.5234, tag: '', tagClass: '' },
	{ name: 'OP', price: 1.32, tag: '', tagClass: '' },
	{ name: 'NEAR', price: 3.12, tag: '', tagClass: '' },
	{ name: 'STX', price: 0.9876, tag: '', tagClass: '' },
	{ name: 'INJ', price: 12.45, tag: '', tagClass: '' },
	{ name: 'CFX', price: 0.1234, tag: '', tagClass: '' },
	{ name: 'FTM', price: 0.4521, tag: '', tagClass: '' },
	{ name: 'HBAR', price: 0.0512, tag: '', tagClass: '' },
	// DeFi
	{ name: 'UNI', price: 5.87, tag: '', tagClass: '' },
	{ name: 'AAVE', price: 168.32, tag: '', tagClass: '' },
	{ name: 'MKR', price: 1456.78, tag: '', tagClass: '' },
	{ name: 'COMP', price: 42.15, tag: '', tagClass: '' },
	{ name: 'SNX', price: 1.23, tag: '', tagClass: '' },
	{ name: 'CRV', price: 0.2845, tag: '', tagClass: '' },
	{ name: 'SUSHI', price: 0.6532, tag: '', tagClass: '' },
	{ name: 'YFI', price: 5234.56, tag: '', tagClass: '' },
	{ name: '1INCH', price: 0.2156, tag: '', tagClass: '' },
	{ name: 'LDO', price: 1.12, tag: '', tagClass: '' },
	{ name: 'GMX', price: 18.76, tag: '', tagClass: '' },
	{ name: 'PENDLE', price: 2.34, tag: '', tagClass: '' },
	// AI概念
	{ name: 'FET', price: 0.4532, tag: 'AI', tagClass: 'ai' },
	{ name: 'AGIX', price: 0.3245, tag: 'AI', tagClass: 'ai' },
	{ name: 'OCEAN', price: 0.3876, tag: 'AI', tagClass: 'ai' },
	{ name: 'RNDR', price: 4.56, tag: 'AI', tagClass: 'ai' },
	{ name: 'WLD', price: 1.78, tag: 'AI', tagClass: 'ai' },
	// 游戏/元宇宙
	{ name: 'SAND', price: 0.2876, tag: '', tagClass: '' },
	{ name: 'MANA', price: 0.2654, tag: '', tagClass: '' },
	{ name: 'AXS', price: 4.32, tag: '', tagClass: '' },
	{ name: 'GALA', price: 0.0187, tag: '', tagClass: '' },
	{ name: 'ILV', price: 28.54, tag: '', tagClass: '' },
	{ name: 'ENJ', price: 0.1234, tag: '', tagClass: '' },
	{ name: 'FLOW', price: 0.4521, tag: '', tagClass: '' },
	{ name: 'CHZ', price: 0.0543, tag: '', tagClass: '' },
	// 基础设施
	{ name: 'GRT', price: 0.1234, tag: '', tagClass: '' },
	{ name: 'FIL', price: 2.87, tag: '', tagClass: '' },
	{ name: 'ICP', price: 5.43, tag: '', tagClass: '' },
	{ name: 'THETA', price: 0.9876, tag: '', tagClass: '' },
	{ name: 'STORJ', price: 0.3456, tag: '', tagClass: '' },
	{ name: 'ROSE', price: 0.0456, tag: '', tagClass: '' },
	{ name: 'LRC', price: 0.1234, tag: '', tagClass: '' },
	// Meme币
	{ name: 'SHIB', price: 0.00001234, tag: 'MEME', tagClass: 'meme' },
	{ name: 'PEPE', price: 0.00000501, tag: 'MEME', tagClass: 'meme' },
	{ name: 'BLUR', price: 0.1876, tag: '', tagClass: '' },
	// 稳定币
	{ name: 'USDT', price: 1.0001, tag: '稳定币', tagClass: 'stable' },
	{ name: 'USDC', price: 0.9999, tag: '稳定币', tagClass: 'stable' },
	{ name: 'FDUSD', price: 0.9988, tag: '0手续费', tagClass: 'free' },
	// 其他
	{ name: 'ALGO', price: 0.1456, tag: '', tagClass: '' },
	{ name: 'EGLD', price: 23.45, tag: '', tagClass: '' },
	{ name: 'XTZ', price: 0.6543, tag: '', tagClass: '' },
	{ name: 'CAKE', price: 1.56, tag: '', tagClass: '' },
	{ name: 'KAVA', price: 0.3245, tag: '', tagClass: '' },
	{ name: 'CELO', price: 0.4123, tag: '', tagClass: '' },
	{ name: 'ZIL', price: 0.0123, tag: '', tagClass: '' },
	{ name: 'ONT', price: 0.1543, tag: '', tagClass: '' },
	{ name: 'WAVES', price: 1.23, tag: '', tagClass: '' },
	{ name: 'QTUM', price: 2.15, tag: '', tagClass: '' },
	{ name: 'IOST', price: 0.0054, tag: '', tagClass: '' },
	{ name: 'IOTX', price: 0.0234, tag: '', tagClass: '' },
	{ name: 'ANKR', price: 0.0187, tag: '', tagClass: '' },
	{ name: 'WOO', price: 0.1234, tag: '', tagClass: '' },
	{ name: 'APE', price: 0.5432, tag: '', tagClass: '' },
	{ name: 'IMX', price: 0.8765, tag: '', tagClass: '' },
	{ name: 'RUNE', price: 2.87, tag: '', tagClass: '' },
	{ name: 'ENS', price: 16.78, tag: '', tagClass: '' },
	{ name: 'KNC', price: 0.3876, tag: '', tagClass: '' },
	{ name: 'NMR', price: 12.34, tag: '', tagClass: '' },
	{ name: 'RLC', price: 1.23, tag: '', tagClass: '' },
	{ name: 'LPT', price: 8.76, tag: '', tagClass: '' },
	{ name: 'MASK', price: 2.15, tag: '', tagClass: '' },
	{ name: 'GLM', price: 0.2345, tag: '', tagClass: '' },
	{ name: 'AUDIO', price: 0.1123, tag: '', tagClass: '' },
	{ name: 'SKL', price: 0.0234, tag: '', tagClass: '' },
	{ name: 'CTSI', price: 0.0987, tag: '', tagClass: '' },
	{ name: 'JASMY', price: 0.0123, tag: '', tagClass: '' },
	{ name: 'ZEC', price: 34.56, tag: '', tagClass: '' },
	{ name: 'DASH', price: 23.45, tag: '', tagClass: '' },
	{ name: 'BAT', price: 0.1543, tag: '', tagClass: '' }
]

/**
 * 获取币种图标路径
 * @param {string} name 币种名称 (如 'BTC', 'ETH')
 * @returns {string} 图标文件的本地路径
 */
export const getCryptoIcon = (name) => {
	if (!name) return '/static/crypto-icons/generic.svg'
	
	const coinName = name.toLowerCase()
	// 合并后的图标库列表 (包含100+主流币种)
	const availableIcons = [
		'1inch', 'aave', 'ada', 'agix', 'algo', 'ankr', 'ape', 'apt', 'arb', 'atom', 
		'audio', 'avax', 'axs', 'bat', 'bch', 'blur', 'bnb', 'btc', 'cake', 'celo', 
		'cfx', 'chz', 'comp', 'crv', 'ctsi', 'dash', 'doge', 'dot', 'egld', 'enj', 
		'ens', 'eos', 'etc', 'eth', 'fdusd', 'fet', 'fil', 'flow', 'ftm', 'gala', 
		'glm', 'gmx', 'grt', 'hbar', 'icp', 'ilv', 'imx', 'inj', 'iost', 'iotx', 
		'jasmy', 'kava', 'knc', 'ldo', 'link', 'lpt', 'lrc', 'ltc', 'mana', 'mask', 
		'matic', 'mkr', 'near', 'neo', 'nmr', 'ocean', 'ont', 'op', 'pendle', 'pepe', 
		'qtum', 'rlc', 'rndr', 'rose', 'rune', 'sand', 'shib', 'skl', 'snx', 'sol', 
		'storj', 'stx', 'sui', 'sushi', 'theta', 'trx', 'uni', 'usdc', 'usdt', 'vet', 
		'waves', 'wld', 'woo', 'xlm', 'xmr', 'xrp', 'xtz', 'yfi', 'zec', 'zil'
	]
	
	if (availableIcons.includes(coinName)) {
		return `/static/crypto-icons/${coinName}.svg`
	}
	
	return '/static/crypto-icons/generic.svg'
}

/**
 * 生成模拟市场数据的通用函数
 */
export const generateMarketData = (baseData) => {
	return baseData.map(item => {
		const changePercent = (Math.random() * 10 - 5).toFixed(2)
		const isUp = parseFloat(changePercent) >= 0
		
		let volumeNum
		if (item.price > 10000) {
			volumeNum = Math.random() * 5000000000 + 1000000000
		} else if (item.price > 100) {
			volumeNum = Math.random() * 2000000000 + 500000000
		} else if (item.price > 1) {
			volumeNum = Math.random() * 800000000 + 100000000
		} else {
			volumeNum = Math.random() * 500000000 + 50000000
		}
		
		let volumeStr
		if (volumeNum >= 1000000000) {
			volumeStr = (volumeNum / 1000000000).toFixed(2) + 'B'
		} else {
			volumeStr = (volumeNum / 1000000).toFixed(2) + 'M'
		}
		
		let priceStr
		if (item.price >= 1000) {
			priceStr = item.price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
		} else if (item.price >= 1) {
			priceStr = item.price.toFixed(4)
		} else if (item.price >= 0.0001) {
			priceStr = item.price.toFixed(6)
		} else {
			priceStr = item.price.toFixed(10)
		}
		
		return {
			name: item.name,
			quote: 'USDT',
			symbol: `${item.name}/USDT`,
			icon: item.name.toLowerCase(),
			tag: item.tag,
			tagClass: item.tagClass,
			volume: volumeStr,
			volumeNum: volumeNum,
			price: priceStr,
			priceNum: item.price,
			usd: priceStr,
			change: `${isUp ? '+' : ''}${changePercent}%`,
			changeNum: parseFloat(changePercent),
			changeClass: isUp ? 'up' : 'down'
		}
	})
}
