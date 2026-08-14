<template>
    <el-form-item :prop="prop" :rules="resolvedRules" :required="required" v-bind="$attrs">
        <!-- Jika menggunakan slot #label -->
        <template v-if="$slots.label" #label>
            <slot name="label" />
        </template>

        <!-- Fallback ke prop label, atau gunakan prop jika label kosong -->
        <template v-else-if="label" #label>
            {{ label }}
        </template>

        <slot />
    </el-form-item>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
    defineProps<{
        label?: string // Diubah menjadi opsional (menggunakan tanda ?)
        prop: string
        required?: boolean
        rules?: any[] | object
    }>(),
    {
        required: false
    }
)

// Otomatis generate pesan error (fallback ke prop jika label string tidak diisi)
const resolvedRules = computed(() => {
    const baseRules: any[] = []
    const fieldName = props.label || props.prop

    if (props.required) {
        baseRules.push({
            required: true,
            message: `${fieldName} wajib diisi`,
            trigger: ['blur', 'change']
        })
    }

    if (props.rules) {
        if (Array.isArray(props.rules)) {
            baseRules.push(...props.rules)
        } else {
            baseRules.push(props.rules)
        }
    }

    return baseRules
})
</script>