package mapper

func ToExcel[Anything any](array []Anything) (filename string) {
	panic("implement me")
	// TODO
	//1.  через reflect получить список полей у структуры которая внутри интерфейса Anything
	//2. В каждой строке элемент структуры из слайса
	//3. в колонках строки поля структуры
	//4. п. 3 в эксель табличку
}
