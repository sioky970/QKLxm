<template>
	<image 
		class="svg-icon"
		:src="iconSrc"
		:style="iconStyle"
		mode="aspectFit"
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
		default: ''
	}
})

const iconPath = computed(() => {
	return `/static/icons/${props.name}.svg`
})

const iconSrc = computed(() => {
	return iconPath.value
})

const iconStyle = computed(() => {
	const sizeValue = typeof props.size === 'number' ? `${props.size}px` : props.size
	return {
		width: sizeValue,
		height: sizeValue
	}
})
</script>

<style scoped lang="scss">
.svg-icon {
	flex-shrink: 0;
	display: inline-block;
}
</style>
