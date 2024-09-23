package report

import "github.com/gocroot/model"

type PhoneNumberInfo struct {
	Count float64
	Name  string
}

// // Merge the counts
// mergedCounts := mergePhoneNumberCounts(pushReportCounts, laporanCounts)

// // Print merged counts
// for phoneNumber, info := range mergedCounts {
// 	fmt.Printf("Phone Number: %s, Count: %d, Name: %s\n", phoneNumber, info.Count, info.Name)
// }
func MergePhoneNumberCounts(counts1, counts2 map[string]PhoneNumberInfo) map[string]PhoneNumberInfo {
	mergedCounts := make(map[string]PhoneNumberInfo)

	for phoneNumber, info1 := range counts1 {
		mergedCounts[phoneNumber] = info1
	}

	for phoneNumber, info2 := range counts2 {
		if info1, exists := mergedCounts[phoneNumber]; exists {
			info1.Count += info2.Count
			mergedCounts[phoneNumber] = info1
		} else {
			mergedCounts[phoneNumber] = info2
		}
	}

	return mergedCounts
}

// phoneNumberCount := countDuplicatePhoneNumbersLaporan(laporans)

// 	for phoneNumber, info := range phoneNumberCount {
// 		fmt.Printf("Phone Number: %s, Count: %d, Name: %s\n", phoneNumber, info.Count, info.Name)
// 	}
func CountDuplicatePhoneNumbersLaporan(laporans []model.Laporan) map[string]PhoneNumberInfo {
	phoneNumberCount := make(map[string]PhoneNumberInfo)

	for _, laporan := range laporans {
		phoneNumber := laporan.User.PhoneNumber
		name := laporan.User.Name
		rating := laporan.Rating / 5 //langsung dibagi 5 untuk dabat bonus rating
		if phoneNumber != "" {
			if info, exists := phoneNumberCount[phoneNumber]; exists {
				info.Count++
				info.Count += rating
				phoneNumberCount[phoneNumber] = info
			} else {
				phoneNumberCount[phoneNumber] = PhoneNumberInfo{Count: 1, Name: name}
			}
		}
	}

	return phoneNumberCount
}

// phoneNumberCount := countDuplicatePhoneNumbers(reports)

// for phoneNumber, info := range phoneNumberCount {
// 	fmt.Printf("Phone Number: %s, Count: %d, Name: %s\n", phoneNumber, info.Count, info.Name)
// }
func CountDuplicatePhoneNumbersWithName(reports []model.PushReport) map[string]PhoneNumberInfo {
	phoneNumberCount := make(map[string]PhoneNumberInfo)

	for _, report := range reports {
		phoneNumber := report.User.PhoneNumber
		if phoneNumber != "" {
			if info, exists := phoneNumberCount[phoneNumber]; exists {
				info.Count++
				phoneNumberCount[phoneNumber] = info
			} else {
				phoneNumberCount[phoneNumber] = PhoneNumberInfo{Count: 1, Name: report.User.Name}
			}
		}
	}

	return phoneNumberCount
}

// phoneNumberCount := countDuplicatePhoneNumbers(reports)

// 	for phoneNumber, count := range phoneNumberCount {
// 		fmt.Printf("Phone Number: %s, Count: %d\n", phoneNumber, count)
// 	}
func CountDuplicatePhoneNumbers(reports []model.PushReport) map[string]int {
	phoneNumberCount := make(map[string]int)

	for _, report := range reports {
		phoneNumber := report.User.PhoneNumber
		if phoneNumber != "" {
			phoneNumberCount[phoneNumber]++
		}
	}

	return phoneNumberCount
}

//emailCount := countDuplicateEmails(reports)
//for email, count := range emailCount {
//		fmt.Printf("Email: %s, Count: %d\n", email, count)
//}
func CountDuplicateEmails(reports []model.PushReport) map[string]int {
	emailCount := make(map[string]int)

	for _, report := range reports {
		if report.Email != "" {
			emailCount[report.Email]++
		}
	}

	return emailCount
}

//projectCount := countDuplicateProjects(reports)
//for project, count := range projectCount {
//  fmt.Printf("Project: %s, Count: %d\n", project, count)
//}
func CountDuplicateProjects(reports []model.PushReport) map[string]int {
	projectCount := make(map[string]int)

	for _, report := range reports {
		projectName := report.Project.Name
		if projectName != "" {
			projectCount[projectName]++
		}
	}

	return projectCount
}
