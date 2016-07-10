function [mat] = combiner_matrix(n)
  mat = sparse(n, n);
  mat(1, 1) = 1;
  mat(1, n/2+1) = 1;
  for i = 2:(n/4)
    mat(i, i) = 1;
    mat(i, i+n/2) = cos(2*pi/n*(i-1));
    mat(i, i+n/2+n/4) = -sin(2*pi/n*(i-1));
  end
  mat(n/4+1, n/4+1) = 1;
  for i = 1:(n/4)
    mirrorRow = n/4 + 1 - i;
    mat(n/4+1+i, :) = mat(mirrorRow, :);
    mat(n/4+1+i, (n/2+1):n) *= -1;
  end
  for i = 1:(n/4-1)
    rowIdx = n/2 + 1 + i;
    mat(rowIdx, n/4+1+i) = 1;
    mat(rowIdx, n/2+1+i) = sin(2*pi/n*(i));
    mat(rowIdx, n/2+n/4+1+i) = cos(2*pi/n*(i));
  end
  mat(n-(n/4-1), n-(n/4-1)) = 1;
  for i = 1:(n/4-1)
    rowIdx = n - n/4 + i + 1;
    mat(rowIdx, :) = mat(n-n/4-i+1, :);
    mat(rowIdx, 1:(n/2)) *= -1;
  end
end
