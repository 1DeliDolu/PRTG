// Büyük veri dosyalarını parçalara bölen script
const fs = require('fs');
const path = require('path');

function splitLargeFile(filePath, chunkSize = 1000) {
    const data = fs.readFileSync(filePath, 'utf8');
    const lines = data.split('\n');
    
    if (lines.length <= chunkSize) {
        console.log('Dosya zaten yeterince küçük');
        return;
    }
    
    const fileName = path.basename(filePath, path.extname(filePath));
    const fileExt = path.extname(filePath);
    const dir = path.dirname(filePath);
    
    for (let i = 0; i < lines.length; i += chunkSize) {
        const chunk = lines.slice(i, i + chunkSize);
        const chunkFileName = `${fileName}_part_${Math.floor(i / chunkSize) + 1}${fileExt}`;
        const chunkPath = path.join(dir, 'chunks', chunkFileName);
        
        // chunks klasörünü oluştur
        fs.mkdirSync(path.join(dir, 'chunks'), { recursive: true });
        
        fs.writeFileSync(chunkPath, chunk.join('\n'));
        console.log(`Oluşturuldu: ${chunkFileName}`);
    }
}

// Kullanım örneği
// splitLargeFile('./data/large_file.csv', 1000);

module.exports = { splitLargeFile };