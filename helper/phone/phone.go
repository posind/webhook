package phone

func MaskPhoneNumber(phone string) string {
	if len(phone) < 9 {
		// Jika nomor telepon terlalu pendek untuk disamarkan, kembalikan tanpa perubahan
		return phone
	}
	// Ambil bagian pertama dari nomor telepon hingga posisi ke-6, tambahkan "xxx", lalu tambahkan sisa dari digit ke-9
	return phone[:6] + "xxx" + phone[9:]
}
