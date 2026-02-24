<template>
	<view class="qrcode-container">
		<canvas 
			:canvas-id="canvasId" 
			:id="canvasId"
			:style="{ width: size + 'px', height: size + 'px' }"
			class="qrcode-canvas"
		/>
	</view>
</template>

<script setup>
import { ref, watch, onMounted, nextTick } from 'vue'

const props = defineProps({
	text: {
		type: String,
		required: true
	},
	size: {
		type: Number,
		default: 200
	},
	margin: {
		type: Number,
		default: 10
	},
	backgroundColor: {
		type: String,
		default: '#ffffff'
	},
	foregroundColor: {
		type: String,
		default: '#000000'
	}
})

const canvasId = ref('qrcode_' + Math.random().toString(36).substr(2, 9))

// 二维码数据矩阵
let modules = []
let moduleCount = 0

// 初始化
onMounted(() => {
	nextTick(() => {
		if (props.text) {
			generateQRCode(props.text)
		}
	})
})

// 监听text变化
watch(() => props.text, (newVal) => {
	if (newVal) {
		nextTick(() => {
			generateQRCode(newVal)
		})
	}
})

// 生成二维码
const generateQRCode = (text) => {
	if (!text) return
	
	try {
		// 创建二维码数据
		const qr = createQRCode(text, 4) // 使用纠错级别M
		modules = qr.modules
		moduleCount = qr.moduleCount
		
		// 绘制二维码
		drawQRCode()
	} catch (err) {
		console.error('QRCode generation error:', err)
	}
}

// 绘制二维码到canvas
const drawQRCode = () => {
	const ctx = uni.createCanvasContext(canvasId.value)
	const size = props.size
	const margin = props.margin
	const cellSize = (size - margin * 2) / moduleCount
	
	// 清空画布并绘制背景
	ctx.setFillStyle(props.backgroundColor)
	ctx.fillRect(0, 0, size, size)
	
	// 绘制二维码模块
	ctx.setFillStyle(props.foregroundColor)
	for (let row = 0; row < moduleCount; row++) {
		for (let col = 0; col < moduleCount; col++) {
			if (modules[row][col]) {
				const x = margin + col * cellSize
				const y = margin + row * cellSize
				ctx.fillRect(x, y, cellSize, cellSize)
			}
		}
	}
	
	ctx.draw()
}

// ================== 二维码生成算法 ==================

// 创建二维码
function createQRCode(text, errorCorrectionLevel) {
	const mode = getMode(text)
	const typeNumber = getTypeNumber(text, errorCorrectionLevel)
	
	const qr = {
		typeNumber,
		errorCorrectionLevel,
		modules: [],
		moduleCount: 0
	}
	
	makeImpl(qr, text, mode)
	
	return qr
}

// 获取模式
function getMode(text) {
	// 简化：全部使用字节模式
	return 4 // BYTE mode
}

// 获取版本号
function getTypeNumber(text, ecl) {
	const length = encodeURIComponent(text).replace(/%[0-9A-F]{2}/g, 'x').length
	// 简化计算
	if (length <= 17) return 1
	if (length <= 32) return 2
	if (length <= 53) return 3
	if (length <= 78) return 4
	if (length <= 106) return 5
	if (length <= 134) return 6
	if (length <= 154) return 7
	if (length <= 192) return 8
	if (length <= 230) return 9
	return 10
}

// 生成二维码实现
function makeImpl(qr, text, mode) {
	qr.moduleCount = qr.typeNumber * 4 + 17
	qr.modules = new Array(qr.moduleCount)
	
	for (let row = 0; row < qr.moduleCount; row++) {
		qr.modules[row] = new Array(qr.moduleCount)
		for (let col = 0; col < qr.moduleCount; col++) {
			qr.modules[row][col] = null
		}
	}
	
	// 设置功能图案
	setupPositionProbePattern(qr, 0, 0)
	setupPositionProbePattern(qr, qr.moduleCount - 7, 0)
	setupPositionProbePattern(qr, 0, qr.moduleCount - 7)
	setupPositionAdjustPattern(qr)
	setupTimingPattern(qr)
	setupTypeInfo(qr, true, 0)
	
	if (qr.typeNumber >= 7) {
		setupTypeNumber(qr, true)
	}
	
	// 获取数据
	const data = createData(qr, text, mode)
	mapData(qr, data, 0)
}

// 定位图案
function setupPositionProbePattern(qr, row, col) {
	for (let r = -1; r <= 7; r++) {
		if (row + r <= -1 || qr.moduleCount <= row + r) continue
		for (let c = -1; c <= 7; c++) {
			if (col + c <= -1 || qr.moduleCount <= col + c) continue
			if ((0 <= r && r <= 6 && (c === 0 || c === 6)) ||
				(0 <= c && c <= 6 && (r === 0 || r === 6)) ||
				(2 <= r && r <= 4 && 2 <= c && c <= 4)) {
				qr.modules[row + r][col + c] = true
			} else {
				qr.modules[row + r][col + c] = false
			}
		}
	}
}

// 对齐图案
function setupPositionAdjustPattern(qr) {
	const pos = getPatternPosition(qr.typeNumber)
	for (let i = 0; i < pos.length; i++) {
		for (let j = 0; j < pos.length; j++) {
			const row = pos[i]
			const col = pos[j]
			if (qr.modules[row][col] !== null) continue
			for (let r = -2; r <= 2; r++) {
				for (let c = -2; c <= 2; c++) {
					if (r === -2 || r === 2 || c === -2 || c === 2 || (r === 0 && c === 0)) {
						qr.modules[row + r][col + c] = true
					} else {
						qr.modules[row + r][col + c] = false
					}
				}
			}
		}
	}
}

// 时序图案
function setupTimingPattern(qr) {
	for (let r = 8; r < qr.moduleCount - 8; r++) {
		if (qr.modules[r][6] !== null) continue
		qr.modules[r][6] = (r % 2 === 0)
	}
	for (let c = 8; c < qr.moduleCount - 8; c++) {
		if (qr.modules[6][c] !== null) continue
		qr.modules[6][c] = (c % 2 === 0)
	}
}

// 类型信息
function setupTypeInfo(qr, test, maskPattern) {
	const data = (2 << 3) | maskPattern // ECL=M, mask=0
	const bits = getBCHTypeInfo(data)
	
	for (let i = 0; i < 15; i++) {
		const mod = (!test && ((bits >> i) & 1) === 1)
		if (i < 6) {
			qr.modules[i][8] = mod
		} else if (i < 8) {
			qr.modules[i + 1][8] = mod
		} else {
			qr.modules[qr.moduleCount - 15 + i][8] = mod
		}
	}
	
	for (let i = 0; i < 15; i++) {
		const mod = (!test && ((bits >> i) & 1) === 1)
		if (i < 8) {
			qr.modules[8][qr.moduleCount - i - 1] = mod
		} else if (i < 9) {
			qr.modules[8][15 - i - 1 + 1] = mod
		} else {
			qr.modules[8][15 - i - 1] = mod
		}
	}
	qr.modules[qr.moduleCount - 8][8] = (!test)
}

// 版本信息
function setupTypeNumber(qr, test) {
	const bits = getBCHTypeNumber(qr.typeNumber)
	for (let i = 0; i < 18; i++) {
		const mod = (!test && ((bits >> i) & 1) === 1)
		qr.modules[Math.floor(i / 3)][i % 3 + qr.moduleCount - 8 - 3] = mod
	}
	for (let i = 0; i < 18; i++) {
		const mod = (!test && ((bits >> i) & 1) === 1)
		qr.modules[i % 3 + qr.moduleCount - 8 - 3][Math.floor(i / 3)] = mod
	}
}

// 创建数据
function createData(qr, text, mode) {
	const buffer = new QRBitBuffer()
	
	// 模式指示符
	buffer.put(4, 4) // BYTE mode
	
	// 字符数
	const bytes = stringToBytes(text)
	buffer.put(bytes.length, 8)
	
	// 数据
	for (let i = 0; i < bytes.length; i++) {
		buffer.put(bytes[i], 8)
	}
	
	// 获取数据容量
	const totalDataCount = getTotalDataCount(qr.typeNumber, 2) // ECL=M
	
	// 终止符
	if (buffer.getLengthInBits() + 4 <= totalDataCount * 8) {
		buffer.put(0, 4)
	}
	
	// 填充到字节边界
	while (buffer.getLengthInBits() % 8 !== 0) {
		buffer.putBit(false)
	}
	
	// 填充码
	while (true) {
		if (buffer.getLengthInBits() >= totalDataCount * 8) break
		buffer.put(0xEC, 8)
		if (buffer.getLengthInBits() >= totalDataCount * 8) break
		buffer.put(0x11, 8)
	}
	
	return createBytes(buffer, qr.typeNumber)
}

// 映射数据
function mapData(qr, data, maskPattern) {
	let inc = -1
	let row = qr.moduleCount - 1
	let bitIndex = 7
	let byteIndex = 0
	
	for (let col = qr.moduleCount - 1; col > 0; col -= 2) {
		if (col === 6) col--
		while (true) {
			for (let c = 0; c < 2; c++) {
				if (qr.modules[row][col - c] === null) {
					let dark = false
					if (byteIndex < data.length) {
						dark = ((data[byteIndex] >>> bitIndex) & 1) === 1
					}
					const mask = getMask(maskPattern, row, col - c)
					if (mask) dark = !dark
					qr.modules[row][col - c] = dark
					bitIndex--
					if (bitIndex === -1) {
						byteIndex++
						bitIndex = 7
					}
				}
			}
			row += inc
			if (row < 0 || qr.moduleCount <= row) {
				row -= inc
				inc = -inc
				break
			}
		}
	}
}

// 遮罩函数
function getMask(maskPattern, row, col) {
	switch (maskPattern) {
		case 0: return (row + col) % 2 === 0
		case 1: return row % 2 === 0
		case 2: return col % 3 === 0
		case 3: return (row + col) % 3 === 0
		case 4: return (Math.floor(row / 2) + Math.floor(col / 3)) % 2 === 0
		case 5: return (row * col) % 2 + (row * col) % 3 === 0
		case 6: return ((row * col) % 2 + (row * col) % 3) % 2 === 0
		case 7: return ((row * col) % 3 + (row + col) % 2) % 2 === 0
	}
	return false
}

// BCH编码
function getBCHTypeInfo(data) {
	let d = data << 10
	while (getBCHDigit(d) - getBCHDigit(0x537) >= 0) {
		d ^= (0x537 << (getBCHDigit(d) - getBCHDigit(0x537)))
	}
	return ((data << 10) | d) ^ 0x5412
}

function getBCHTypeNumber(data) {
	let d = data << 12
	while (getBCHDigit(d) - getBCHDigit(0x1F25) >= 0) {
		d ^= (0x1F25 << (getBCHDigit(d) - getBCHDigit(0x1F25)))
	}
	return (data << 12) | d
}

function getBCHDigit(data) {
	let digit = 0
	while (data !== 0) {
		digit++
		data >>>= 1
	}
	return digit
}

// 对齐图案位置
function getPatternPosition(typeNumber) {
	const PATTERN_POSITION_TABLE = [
		[],
		[6, 18],
		[6, 22],
		[6, 26],
		[6, 30],
		[6, 34],
		[6, 22, 38],
		[6, 24, 42],
		[6, 26, 46],
		[6, 28, 50],
		[6, 30, 54],
	]
	return PATTERN_POSITION_TABLE[typeNumber] || []
}

// 数据容量
function getTotalDataCount(typeNumber, errorCorrectionLevel) {
	const RS_BLOCK_TABLE = [
		[1, 26, 19],
		[1, 44, 34],
		[1, 70, 55],
		[1, 100, 80],
		[1, 134, 108],
		[2, 86, 68],
		[2, 98, 78],
		[2, 121, 97],
		[2, 146, 116],
		[2, 86, 68, 2, 87, 69],
	]
	const blocks = RS_BLOCK_TABLE[typeNumber - 1]
	if (!blocks) return 0
	
	let total = 0
	for (let i = 0; i < blocks.length; i += 3) {
		total += blocks[i] * blocks[i + 2]
	}
	return total
}

// 创建最终字节
function createBytes(buffer, typeNumber) {
	const totalCount = getTotalDataCount(typeNumber, 2)
	const data = new Array(totalCount)
	for (let i = 0; i < totalCount; i++) {
		data[i] = buffer.buffer[i] || 0
	}
	return data
}

// 字符串转字节
function stringToBytes(str) {
	const bytes = []
	for (let i = 0; i < str.length; i++) {
		let c = str.charCodeAt(i)
		if (c < 0x80) {
			bytes.push(c)
		} else if (c < 0x800) {
			bytes.push(0xc0 | (c >> 6))
			bytes.push(0x80 | (c & 0x3f))
		} else {
			bytes.push(0xe0 | (c >> 12))
			bytes.push(0x80 | ((c >> 6) & 0x3f))
			bytes.push(0x80 | (c & 0x3f))
		}
	}
	return bytes
}

// BitBuffer类
class QRBitBuffer {
	constructor() {
		this.buffer = []
		this.length = 0
	}
	
	get(index) {
		const bufIndex = Math.floor(index / 8)
		return ((this.buffer[bufIndex] >>> (7 - index % 8)) & 1) === 1
	}
	
	put(num, length) {
		for (let i = 0; i < length; i++) {
			this.putBit(((num >>> (length - i - 1)) & 1) === 1)
		}
	}
	
	getLengthInBits() {
		return this.length
	}
	
	putBit(bit) {
		const bufIndex = Math.floor(this.length / 8)
		if (this.buffer.length <= bufIndex) {
			this.buffer.push(0)
		}
		if (bit) {
			this.buffer[bufIndex] |= (0x80 >>> (this.length % 8))
		}
		this.length++
	}
}
</script>

<style scoped lang="scss">
.qrcode-container {
	display: flex;
	justify-content: center;
	align-items: center;
}

.qrcode-canvas {
	background: #fff;
}
</style>
