<template>
	<view 
		class="svg-icon" 
		:style="iconStyle"
	/>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
	name: {
		type: String,
		required: true
	},
	size: {
		type: [Number, String],
		default: 24
	},
	color: {
		type: String,
		default: '#000000'
	}
})

// 图标路径
const iconPath = computed(() => `/static/icons/${props.name}.svg`)

// 图标样式
const iconStyle = computed(() => {
	const sizeValue = typeof props.size === 'number' ? `${props.size}px` : props.size
	return {
		width: sizeValue,
		height: sizeValue,
		backgroundColor: props.color,
		maskImage: `url(${iconPath.value})`,
		webkitMaskImage: `url(${iconPath.value})`,
		maskSize: 'contain',
		webkitMaskSize: 'contain',
		maskRepeat: 'no-repeat',
		webkitMaskRepeat: 'no-repeat',
		display: 'inline-block'
	}
})
</script>

<style scoped lang="scss">
.svg-icon {
	flex-shrink: 0;
}
</style>
